package broker

import (
	"context"
	"encoding/json"
	"github.com/IsaacDSC/event-driven/internal/utils"
	"github.com/IsaacDSC/event-driven/types"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"time"
)

type PublisherServer struct {
	client *asynq.Client
}

var _ types.Producer = (*PublisherServer)(nil)

func NewProducerServer(addr string) *PublisherServer {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: addr})

	return &PublisherServer{client: client}
}

func (ps PublisherServer) Close() {
	ps.client.Close()
}

func (ps PublisherServer) Producer(ctx context.Context, input types.PayloadType) error {
	if input.Type != types.EventTypeTask && input.Type != types.EventTypeSaga {
		return nil
	}

	inputTask, err := json.Marshal(input.Payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(input.EventName, inputTask)
	var opts []asynq.Option

	opts = append(opts, asynq.TaskID(input.EventID.String()))
	opts = append(opts, asynq.MaxRetry(input.Opts.MaxRetry))

	if len(input.Opts.Queue) > 0 {
		opts = append(opts, asynq.Queue(input.Opts.Queue))
	}

	if !(input.Opts.Delay == 0) {
		opts = append(opts, asynq.ProcessIn(input.Opts.Delay))
	}

	if !(input.Opts.Unique == 0) {
		opts = append(opts, asynq.Unique(input.Opts.Unique))
	}

	if !(input.Opts.MaxTimeOfProcessTask == 0) {
		opts = append(opts, asynq.Deadline(time.Now().Add(input.Opts.MaxTimeOfProcessTask)))
	}

	if !(input.Opts.DeadLine == time.Time{}) {
		opts = append(opts, asynq.Deadline(input.Opts.DeadLine))
	}

	ctx = utils.SetTxIDToCtx(ctx, input.EventID.String())
	_, err = ps.client.EnqueueContext(ctx, task, opts...)
	if err != nil { //TODO: add sent the configs
		return err
	}

	return nil
}

func (ps PublisherServer) GenerateEventID() uuid.UUID {
	return uuid.New()
}
