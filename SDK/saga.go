package SDK

import (
	"context"
	"fmt"
	"github.com/IsaacDSC/event-driven/internal/broker"
	"github.com/IsaacDSC/event-driven/internal/utils"
	"github.com/IsaacDSC/event-driven/types"
	"github.com/google/uuid"
	"time"
)

type SagaPattern struct {
	ConsumerServer

	Consumers        []types.ConsumerInput
	Options          types.Opts
	SequencePayloads bool

	repository types.Repository
	pb         types.Producer
	log        types.Logger
}

func NewSagaPattern(rdAddr string, repo types.Repository, consumers []types.ConsumerInput, options types.Opts, sequencePayloads bool) *SagaPattern {
	pb := broker.NewProducerServer(rdAddr)

	return &SagaPattern{
		Consumers:        consumers,
		Options:          options,
		SequencePayloads: sequencePayloads,
		repository:       repo,
		pb:               pb,
		log:              utils.NewLogger("[*] - [SagaPattern] - "),
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
				if err := sp.repository.SagaUpdateInfos(ctx, payload.EventID, sp.Options.MaxRetry, "COMMITED_ERROR"); err != nil {
					sp.log.Error("could not save commited_error", utils.KeyLogError.String(), err.Error())
				}
			}

			sp.log.Error("could not execute upFn with error", utils.KeyLogError.String(), err.Error())
			hasError = true
			break
		}

		if sp.repository != nil {
			if err := sp.repository.SagaUpdateInfos(ctx, eventID, 0, "COMMITED"); err != nil {
				sp.log.Error("could not save commited info", utils.KeyLogError.String(), err.Error())
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
						sp.log.Error("could not save backward_error", utils.KeyLogError.String(), err.Error())
					}
				}
				return fmt.Errorf("could not rollback with error: %v", err)
			}

			if sp.repository != nil {
				if err := sp.repository.SagaUpdateInfos(ctx, eventID, configs.MaxRetry, "BACKWARD"); err != nil {
					sp.log.Error("could not save backward info", utils.KeyLogError.String(), err.Error())
				}
			}
		}
	}

	return nil
}

func (sp SagaPattern) executeUpFn(ctx context.Context, fn types.Fn, payload types.PayloadInput, opts types.Opts, attempt int) (err error) {
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

func (sp SagaPattern) executeDownFn(ctx context.Context, fn types.Fn, payload types.PayloadInput, opts types.Opts, attempt int) (err error) {
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
