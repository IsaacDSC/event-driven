package broker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/IsaacDSC/event-driven/internal/sqlc"
	genrepo "github.com/IsaacDSC/event-driven/internal/sqlc/generated/repository"
	"github.com/IsaacDSC/event-driven/types"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"log"
	"time"
)

type SubscriberServer struct {
	server     *asynq.Server
	repository *genrepo.Queries
	mux        *asynq.ServeMux
}

func NewSubscriberServer(addr string, db *sql.DB) *SubscriberServer {
	s := new(SubscriberServer)
	s.repository = sqlc.NewRepository(db)
	s.server = asynq.NewServer(
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
	return s
}

func (ss *SubscriberServer) AddHandlers(consumers map[string]types.ConsumerFn) *SubscriberServer {
	mux := asynq.NewServeMux()
	mux.Use(ss.middleware)
	for eventName, fn := range consumers {
		mux.HandleFunc(eventName, ss.handler(fn))
	}
	ss.mux = mux
	return ss
}

func (ss *SubscriberServer) Start() error {
	if err := ss.server.Run(ss.mux); err != nil {
		return fmt.Errorf("could not run server: %v", err)
	}

	return nil
}

func (ss *SubscriberServer) middleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		start := time.Now()
		log.Printf("Start processing %q", t.Type())

		taskID, _ := asynq.GetTaskID(ctx)
		txID, err := uuid.Parse(taskID)
		if err != nil {
			return err
		}

		if err := h.ProcessTask(ctx, t); err != nil {
			retry, _ := asynq.GetRetryCount(ctx)
			ss.updateInfos(ctx, txID, retry, "ERROR")
			return err
		}

		retry, _ := asynq.GetRetryCount(ctx)
		ss.updateInfos(ctx, txID, retry, "FINISHED")

		log.Printf("Finished processing %q: Elapsed Time = %v", t.Type(), time.Since(start))
		return nil
	})
}

func (ss *SubscriberServer) handler(fn types.ConsumerFn) func(ctx context.Context, t *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var input types.PayloadInput
		if err := json.Unmarshal(t.Payload(), &input); err != nil {
			return err
		}

		if err := fn(ctx, input); err != nil {
			return err
		}

		return nil
	}
}

func (ss *SubscriberServer) updateInfos(ctx context.Context, txID uuid.UUID, retry int, status string) {
	var finishedAt time.Time
	if status == "FINISHED" {
		finishedAt = time.Now()
	}

	if err := ss.repository.UpdateTransaction(ctx, genrepo.UpdateTransactionParams{
		EventID: txID,
		Status:  status,
		TotalRetry: sql.NullInt32{
			Int32: int32(retry),
			Valid: true,
		},
		EndedAt: sql.NullTime{
			Time:  finishedAt,
			Valid: true,
		},
	}); err != nil {
		log.Printf("could not update transaction with error: %v\n", err)
	}

}
