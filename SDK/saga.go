package SDK

import (
	"context"
	"event-driven/internal/acl"
	"event-driven/internal/utils"
	"event-driven/types"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type Fn func(ctx context.Context, payload any, opts ...types.Opts) error

type ConsumerInput interface {
	UpFn(ctx context.Context, payload any, opts ...types.Opts) error
	DownFn(ctx context.Context, payload any, opts ...types.Opts) error
	GetConfig() types.Opts
	GetEventName() string
}

type SagaPattern struct {
	Consumers        []ConsumerInput
	Options          types.Opts
	SequencePayloads bool
	client           *acl.Client
}

func NewSagaPattern(consumers []ConsumerInput, options types.Opts, sequencePayloads bool) *SagaPattern {
	client := acl.NewClient("http://localhost:3333/saga")
	return &SagaPattern{
		Consumers:        consumers,
		Options:          options,
		SequencePayloads: sequencePayloads,
		client:           client,
	}
}

func (sp SagaPattern) Consumer(ctx context.Context, payload map[string]any) error {
	committed := 0
	var hasError bool

	txID, err := utils.GetTxIDFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("could not get txID from context with error: %v", err)
	}

	events := make(map[string]uuid.UUID)

	for _, c := range sp.Consumers {
		configs := sp.getConfig(sp.Options, c.GetConfig())
		eventID := uuid.New()
		events[c.GetEventName()] = eventID

		if err := sp.client.CreateMsg(ctx, types.PayloadType{
			TransactionEventID: txID,
			EventID:            eventID,
			Payload:            payload,
			EventName:          c.GetEventName(),
			Opts:               configs,
			CreatedAt:          time.Now(),
		}); err != nil {
			fmt.Printf("could not create message with error: %v\n", err)
		}

		if err := sp.executeUpFn(ctx, c.UpFn, payload, sp.Options, 0); err != nil {
			hasError = true
			break
		}

		sp.client.UpdateInfos(ctx, eventID, 0, "COMMITED")

		committed++
	}

	if hasError {
		rollbackConsumers := sp.Consumers[:committed+1]
		for i := range rollbackConsumers {
			eventID := events[rollbackConsumers[i].GetEventName()]
			configs := sp.getConfig(sp.Options, rollbackConsumers[i].GetConfig())
			if err := sp.executeDownFn(ctx, rollbackConsumers[i].DownFn, payload, sp.Options, 2); err != nil {
				sp.client.UpdateInfos(ctx, eventID, configs.MaxRetry, "BACKWARD_ERROR")
				return fmt.Errorf("could not rollback with error: %v", err)
			}
			sp.client.UpdateInfos(ctx, eventID, configs.MaxRetry, "BACKWARD")
		}
	}

	return nil
}

func (sp SagaPattern) executeUpFn(ctx context.Context, fn Fn, payload any, opts types.Opts, attempt int) (err error) {
	if opts.MaxRetry == 0 {
		return fn(ctx, payload, opts)
	}

	if err = fn(ctx, payload, sp.Options); err != nil {
		opts.MaxRetry -= 1
		attempt += 1
		backoffDuration := time.Duration(attempt*attempt) * time.Second
		time.Sleep(backoffDuration)
		//	TODO: add registry error in database (retry X with error Y)
		return sp.executeUpFn(ctx, fn, payload, opts, attempt)
	}

	return
}

func (sp SagaPattern) executeDownFn(ctx context.Context, fn Fn, payload any, opts types.Opts, attempt int) (err error) {
	if opts.MaxRetry == 0 {
		return fn(ctx, payload, opts)
	}

	if err = fn(ctx, payload, sp.Options); err != nil {
		opts.MaxRetry -= 1
		attempt += 1
		backoffDuration := time.Duration(attempt*attempt) * time.Second
		time.Sleep(backoffDuration)
		//	TODO: add registry error in database (retry X with error Y)
		return sp.executeDownFn(ctx, fn, payload, opts, attempt)

	}

	return
}

func (sp SagaPattern) getConfig(taskConfig, sagaConfig types.Opts) types.Opts {
	if sagaConfig == types.EmptyOpts {
		return taskConfig
	}

	return sagaConfig

}
