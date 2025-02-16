package SDK

import (
	"context"
	"encoding/json"
	"errors"
	"event-driven/internal/acl"
	"event-driven/internal/utils"
	"event-driven/types"
	"fmt"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"log"
	"time"
)

type ConsumerServer struct {
	server *asynq.Server
	mux    *asynq.ServeMux
	client *acl.Client
}

const baseUrl = "http://localhost:3333/task"

func NewConsumerServer(addr string) *ConsumerServer {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: addr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	return &ConsumerServer{
		server: srv,
		client: acl.NewClient(baseUrl),
	}
}

func (cs *ConsumerServer) AddHandlers(consumers map[string]types.ConsumerFn) *ConsumerServer {
	mux := asynq.NewServeMux()
	mux.Use(cs.middleware)
	for eventName, fn := range consumers {
		mux.HandleFunc(eventName, cs.handler(fn))
	}
	cs.mux = mux
	return cs
}

func (cs *ConsumerServer) handler(fn types.ConsumerFn) func(ctx context.Context, t *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var input map[string]any
		if err := json.Unmarshal(t.Payload(), &input); err != nil {
			return err
		}

		txID, err := utils.GetTxIDFromCtx(ctx)
		if err != nil {
			return err
		}

		if txID == uuid.Nil {
			panic("txID is nil")
		}

		if err := fn(ctx, txID, input); err != nil {
			return err
		}

		return nil
	}
}

func (cs *ConsumerServer) Start() error {
	if err := cs.server.Run(cs.mux); err != nil {
		return fmt.Errorf("could not run server: %v", err)
	}

	return nil
}

func (cs *ConsumerServer) middleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		start := time.Now()
		log.Printf("Start processing %q", t.Type())

		taskID, ok := asynq.GetTaskID(ctx)
		if !ok {
			log.Printf("could not get task ID")
			return errors.New("could not get task ID")
		}

		txID, err := uuid.Parse(taskID)
		if err != nil {
			return err
		}

		ctx = utils.SetTxIDToCtx(ctx, txID)

		if err := h.ProcessTask(ctx, t); err != nil {
			retry, _ := asynq.GetRetryCount(ctx)
			cs.client.UpdateInfos(ctx, txID, retry, "ERROR")
			return err
		}

		retry, _ := asynq.GetRetryCount(ctx)
		cs.client.UpdateInfos(ctx, txID, retry, "FINISHED")

		log.Printf("Finished processing %q: Elapsed Time = %v", t.Type(), time.Since(start))
		return nil
	})
}
