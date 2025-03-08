package SDK

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IsaacDSC/event-driven/internal/utils"
	"github.com/IsaacDSC/event-driven/types"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"time"
)

type ConsumerServer struct {
	server     *asynq.Server
	mux        *asynq.ServeMux
	repository types.Repository
	log        types.Logger
}

func NewConsumerServer(rdAddr string, repo types.Repository) *ConsumerServer {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: rdAddr},
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
		server:     srv,
		repository: repo,
		log:        utils.NewLogger("consumer"),
	}
}

func (cs *ConsumerServer) WithConsumerServer(rdAddr string, repo types.Repository) *ConsumerServer {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: rdAddr},
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
		server:     srv,
		repository: repo,
		log:        utils.NewLogger("[*] - [Consumer] - "),
	}
}

func (cs *ConsumerServer) AddHandlers(consumers map[string]types.ConsumerFn) *ConsumerServer {
	//TODO: Separate for other layers //This Layer not communicate with asynq
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
		var input types.PayloadInput
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

		if err := fn(ctx, input); err != nil {
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
		cs.log.Info("Start processing", utils.KeyAsynqTypeTask.String(), t.Type())

		taskID, ok := asynq.GetTaskID(ctx)
		if !ok {
			cs.log.Error("could not get task ID")
			return errors.New("could not get task ID")
		}

		txID, err := uuid.Parse(taskID)
		if err != nil {
			return err
		}

		ctx = utils.SetTxIDToCtx(ctx, txID)

		if err := h.ProcessTask(ctx, t); err != nil {
			retry, _ := asynq.GetRetryCount(ctx)

			cs.log.Warn("could not process task",
				utils.KeyAsynqTypeTask.String(), t.Type(),
				utils.KeyAsynqTaskID.String(), taskID,
				utils.KeyLogError.String(), err.Error(),
				utils.KeyAsynqRetry.String(), retry,
			)

			if cs.repository != nil {
				if err := cs.repository.UpdateInfos(ctx, txID, retry, "ERROR"); err != nil {
					cs.log.Error("could not save error",
						utils.KeyAsynqTypeTask.String(), t.Type(),
						utils.KeyLogError.String(), err.Error(),
						utils.KeyAsynqTaskID.String(), taskID,
					)
					return err
				}
			}
			return err
		}

		retry, _ := asynq.GetRetryCount(ctx)
		if cs.repository != nil {
			if err := cs.repository.UpdateInfos(ctx, txID, retry, "FINISHED"); err != nil {
				cs.log.Error("could not save finished",
					utils.KeyAsynqTypeTask.String(), t.Type(),
					utils.KeyLogError.String(), err.Error(),
					utils.KeyAsynqTaskID.String(), taskID,
				)
				return err
			}
		}

		cs.log.Info("Finished processing: ",
			utils.KeyAsynqTaskID.String(), taskID,
			utils.KeyAsynqTypeTask.String(), t.Type(),
			utils.KeyAsynqElapsed.String(), time.Since(start),
		)
		return nil
	})
}
