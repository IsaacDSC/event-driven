package SDK

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IsaacDSC/event-driven/internal/broker"
	"github.com/IsaacDSC/event-driven/internal/utils"
	"github.com/IsaacDSC/event-driven/types"
)

type Producer struct {
	host        string
	defaultOpts *types.Opts
	repository  types.Repository
	pb          types.Producer
	dt          types.TimeProvider
	log         types.Logger
}

func NewProducer(rdAddr string, repo types.Repository, defaultOpts *types.Opts) *Producer {
	pb := broker.NewProducerServer(rdAddr)
	dt := utils.NewDateTime()

	if defaultOpts == nil {
		defaultOpts = &types.Opts{
			MaxRetry: 10,
		}
	}

	return &Producer{
		repository:  repo,
		defaultOpts: defaultOpts,
		pb:          pb,
		dt:          dt,
		log:         utils.NewLogger("[*] - [Producer] - "),
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

	eventID := p.pb.GenerateEventID()
	ctx = utils.SetTxIDToCtx(ctx, eventID)
	input := types.PayloadType{
		EventID:     eventID,
		Payload:     inputPayload,
		EventName:   eventName,
		EventsNames: nil,     //TODO: not implemented
		Opts:        opts[0], //TODO: not implemented
		CreatedAt:   p.dt.Now(),
		Type:        eventType,
	}

	if p.repository != nil {
		if err := p.repository.SaveTx(ctx, input); err != nil {
			p.log.Error("could not save transaction", utils.KeyLogError, err.Error())
			return fmt.Errorf("could not create message: %v", err)
		}
	}

	if err := p.pb.Producer(ctx, input); err != nil {
		p.log.Error("could not send message to producer", utils.KeyLogError, err.Error())
		return fmt.Errorf("could not send message with error: %v\n", err)
	}

	return nil
}

func (p Producer) anyToMap(input any) (types.PayloadInput, error) {
	var output types.PayloadInput

	b, err := json.Marshal(input)
	if err != nil {
		return output, fmt.Errorf("could not marshal payload: %v", err)
	}

	output.EventID = p.pb.GenerateEventID()
	output.CreatedAt = p.dt.Now()
	output.Data = b

	return output, nil
}

func (p Producer) Close() {
	p.pb.Close()
}
