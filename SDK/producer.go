package SDK

import (
	"context"
	"database/sql"
	"encoding/json"
	"event-driven/broker"
	"event-driven/database"
	genrepo "event-driven/internal/sqlc/generated/repository"
	"event-driven/internal/utils"
	"event-driven/repository"
	"event-driven/types"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"
)

type Producer struct {
	host        string
	defaultOpts *types.Opts
	repository  types.Repository
	pb          *broker.PublisherServer
	db          *sql.DB
}

func NewProducer(conn types.Connection, defaultOpts *types.Opts) *Producer {
	db, err := database.NewConnection(conn.Database)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	orm := genrepo.New(db)
	repo := repository.NewTransaction(orm)
	pb := broker.NewProducerServer(conn.RedisAddr)

	if defaultOpts == nil {
		defaultOpts = &types.Opts{
			MaxRetry: 10,
		}
	}

	return &Producer{
		repository:  repo,
		defaultOpts: defaultOpts,
		pb:          pb,
		db:          db,
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

	if err := p.repository.SaveTx(ctx, input); err != nil {
		return fmt.Errorf("could not create message: %v", err)
	}

	if err := p.pb.Producer(ctx, input); err != nil {
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

	output.EventID = uuid.New()
	output.CreatedAt = time.Now()
	output.Data = b

	return output, nil
}

func (p Producer) Close() {
	p.pb.Close()
	p.db.Close()
}
