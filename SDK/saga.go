package SDK

import (
	"context"
	"event-driven/internal/broker"
	"event-driven/internal/utils"
	"event-driven/types"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type Fn func(ctx context.Context, payload types.PayloadInput) error

type ConsumerInput interface {
	UpFn(ctx context.Context, payload types.PayloadInput) error
	DownFn(ctx context.Context, payload types.PayloadInput) error
	GetConfig() types.Opts
	GetEventName() string
}

type SagaPattern struct {
	Consumers        []ConsumerInput
	Options          types.Opts
	SequencePayloads bool

	repository types.Repository
	pb         *broker.PublisherServer
}

func NewSagaPattern(rdAddr string, repo types.Repository, consumers []ConsumerInput, options types.Opts, sequencePayloads bool) *SagaPattern {
	pb := broker.NewProducerServer(rdAddr)

	return &SagaPattern{
		Consumers:        consumers,
		Options:          options,
		SequencePayloads: sequencePayloads,
		repository:       repo,
		pb:               pb,
	}
}

func (sp SagaPattern) Consumer(ctx context.Context, payload types.PayloadInput) error {
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

		if sp.repository != nil {
			if err := sp.repository.SagaSaveTx(ctx, types.PayloadType{
				TransactionEventID: txID,
				EventID:            eventID,
				Payload:            payload,
				EventName:          c.GetEventName(),
				Opts:               configs,
				CreatedAt:          time.Now(),
			}); err != nil {
				fmt.Printf("could not create message with error: %v\n", err)
			}
		}

		payload.EventID = eventID
		if err := sp.executeUpFn(ctx, c.UpFn, payload, sp.Options, 0); err != nil {
			if sp.repository != nil {
				sp.repository.SagaUpdateInfos(ctx, payload.EventID, sp.Options.MaxRetry, "COMMITED_ERROR")
			}
			fmt.Printf("could not execute upFn with error: %v\n", err)
			hasError = true
			break
		}

		if sp.repository != nil {
			if err := sp.repository.SagaUpdateInfos(ctx, eventID, 0, "COMMITED"); err != nil {
				fmt.Printf("could not update infos with error: %v\n", err)
			}
		}

		committed++
	}

	if hasError {
		rollbackConsumers := sp.Consumers[:committed+1]
		for i := range rollbackConsumers {
			eventID := events[rollbackConsumers[i].GetEventName()]
			configs := sp.getConfig(sp.Options, rollbackConsumers[i].GetConfig())
			if err := sp.executeDownFn(ctx, rollbackConsumers[i].DownFn, payload, configs, 2); err != nil {
				if sp.repository != nil {
					if err := sp.repository.SagaUpdateInfos(ctx, eventID, configs.MaxRetry, "BACKWARD_ERROR"); err != nil {
						fmt.Printf("error on update info: %v\n", err)
					}
				}
				return fmt.Errorf("could not rollback with error: %v", err)
			}

			if sp.repository != nil {
				if err := sp.repository.SagaUpdateInfos(ctx, eventID, configs.MaxRetry, "BACKWARD"); err != nil {
					fmt.Printf("error on update info: %v\n", err)
				}
			}
		}
	}

	return nil
}

func (sp SagaPattern) executeUpFn(ctx context.Context, fn Fn, payload types.PayloadInput, opts types.Opts, attempt int) (err error) {
	if opts.MaxRetry == 0 {
		return fn(ctx, payload)
	}

	if err = fn(ctx, payload); err != nil {
		opts.MaxRetry -= 1
		attempt += 1
		backoffDuration := time.Duration(attempt*attempt) * time.Second
		time.Sleep(backoffDuration)
		return sp.executeUpFn(ctx, fn, payload, opts, attempt)
	}

	return
}

func (sp SagaPattern) executeDownFn(ctx context.Context, fn Fn, payload types.PayloadInput, opts types.Opts, attempt int) (err error) {
	if opts.MaxRetry == 0 {
		return fn(ctx, payload)
	}

	if err = fn(ctx, payload); err != nil {
		opts.MaxRetry -= 1
		attempt += 1
		backoffDuration := time.Duration(attempt*attempt) * time.Second
		time.Sleep(backoffDuration)
		return sp.executeDownFn(ctx, fn, payload, opts, attempt)
	}

	return
}

func (sp SagaPattern) getConfig(taskConfig, sagaConfig types.Opts) types.Opts {
	if sagaConfig == *types.EmptyOpts {
		return taskConfig
	}

	return sagaConfig

}

func (sp SagaPattern) Close() {
	sp.pb.Close()
}
