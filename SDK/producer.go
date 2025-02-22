package SDK

import (
	"context"
	"encoding/json"
	"event-driven/internal/acl"
	"event-driven/internal/utils"
	"event-driven/types"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type Producer struct {
	host        string
	defaultOpts *types.Opts
	client      *acl.Client
}

func NewProducer(host string, defaultOpts *types.Opts) *Producer {
	if defaultOpts == nil {
		defaultOpts = &types.Opts{
			MaxRetry: 10,
		}
	}
	return &Producer{
		client:      acl.NewClient("http://localhost:3333/task"),
		host:        host,
		defaultOpts: defaultOpts,
	}
}

func (p Producer) SagaProducer(ctx context.Context, eventName string, payload any, opts ...types.Opts) error {
	return p.createMsg(ctx, types.EventTypeSaga, eventName, payload, nil, opts...)
}

func (p Producer) Producer(ctx context.Context, eventName string, payload any, fn types.ConsumerFn, opts ...types.Opts) error {
	return p.createMsg(ctx, types.EventTypeTask, eventName, payload, fn, opts...)
}

func (p Producer) createMsg(ctx context.Context, eventType types.EventType, eventName string, payload any, fn types.ConsumerFn, opts ...types.Opts) error {
	if len(opts) == 0 {
		opts = append(opts, *p.defaultOpts)
	}

	inputPayload, err := p.anyToMap(payload)
	if err != nil {
		return err
	}

	eventID := uuid.New()
	ctx = utils.SetTxIDToCtx(ctx, eventID)
	input := types.PayloadType{
		EventID:     eventID,
		Payload:     inputPayload,
		EventName:   eventName,
		EventsNames: nil,     //TODO: not implemented
		Opts:        opts[0], //TODO: not implemented
		CreatedAt:   time.Now(),
		Type:        eventType,
	}

	if err := p.client.CreateMsg(ctx, input); err != nil {
		return fmt.Errorf("could not create message: %v", err)
	}

	return nil
}

func (p Producer) anyToMap(input any) (types.PayloadInput, error) {
	var output types.PayloadInput

	b, err := json.Marshal(input)
	if err != nil {
		return output, fmt.Errorf("could not marshal payload: %v", err)
	}

	output.EventID = uuid.New()
	output.CreatedAt = time.Now()
	output.Data = b

	return output, nil
}
