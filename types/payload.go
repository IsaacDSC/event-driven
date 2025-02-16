package types

import (
	"context"
	"github.com/google/uuid"
	"time"
)

// Opts
type Opts struct {
	Kind                 string        `json:"kind"`
	MaxRetry             int           `json:"retry"`
	Delay                time.Duration `json:"timout"`
	MaxTimeOfProcessTask time.Duration `json:"max_time_of_process_task"`
	DeadLine             time.Time     `json:"dead_line"`
	Queue                string        `json:"queue"`
	Unique               time.Duration `json:"unique"`
}

type PayloadType struct {
	TransactionEventID uuid.UUID      `json:"transaction_event_id"`
	EventID            uuid.UUID      `json:"event_id"`
	Payload            map[string]any `json:"payload"`
	EventName          string         `json:"event_name"`
	EventsNames        []string       `json:"events_names"`
	Opts               Opts           `json:"opts"`
	Info               map[string]any `json:"info"`
	CreatedAt          time.Time      `json:"created_at"`
}

type UpdatePayloadInput struct {
	Status     string    `json:"status"`
	TotalRetry int       `json:"total_retry"`
	FinishedAt time.Time `json:"finished_at"`
	//Info       map[string]any `json:"info"`
}

// producer
type ProducerFn func(ctx context.Context) error

// consumer
type ConsumerFn func(ctx context.Context, payload map[string]any) error
