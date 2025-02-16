package SDK

import (
	"context"
	"event-driven/internal/acl"
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

func (sp SagaPattern) Consumer(ctx context.Context, txID uuid.UUID, payload map[string]any) error {
	committed := 0
	var hasError bool

	for _, c := range sp.Consumers {
		if err := sp.client.CreateMsg(ctx, types.PayloadType{
			TransactionID: txID,
			EventID:       uuid.New(),
			Payload:       payload,
			EventName:     c.GetEventName(),
			Opts:          c.GetConfig(),
			CreatedAt:     time.Now(),
		}); err != nil {
			return fmt.Errorf("could not create message with error: %v", err)
		}

		if err := sp.executeUpFn(ctx, c.UpFn, payload, sp.Options); err != nil {
			hasError = true
			break
		}
		committed++
	}

	if hasError {
		rollbackConsumers := sp.Consumers[:committed+1]
		// rollback
		for i := range rollbackConsumers {
			if err := sp.executeDownFn(ctx, rollbackConsumers[i].DownFn, payload, sp.Options); err != nil {
				// log error
				//	TODO: add registry error in database (rollback X with error Y)
			}
		}
	}

	return nil
}

func (sp SagaPattern) executeUpFn(ctx context.Context, fn Fn, payload any, opts types.Opts) (err error) {
	if opts.MaxRetry == 0 {
		return fn(ctx, payload, opts)
	}

	if err = fn(ctx, payload, sp.Options); err != nil {
		opts.MaxRetry -= 1
		//	TODO: add registry error in database (retry X with error Y)
		sp.executeUpFn(ctx, fn, payload, opts)

	}

	return
}

func (sp SagaPattern) executeDownFn(ctx context.Context, fn Fn, payload any, opts types.Opts) (err error) {
	if opts.MaxRetry == 0 {
		return fn(ctx, payload, opts)
	}

	if err = fn(ctx, payload, sp.Options); err != nil {
		opts.MaxRetry -= 1
		//	TODO: add registry error in database (retry X with error Y)
		sp.executeDownFn(ctx, fn, payload, opts)

	}

	return
}
