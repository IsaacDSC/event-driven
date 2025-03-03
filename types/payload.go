package types

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

var EmptyOpts = &Opts{}

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

type EventType string

const (
	EventTypeSaga EventType = "saga"
	EventTypeTask EventType = "task"
)

type PayloadType struct {
	TransactionEventID uuid.UUID      `json:"transaction_event_id"`
	EventID            uuid.UUID      `json:"event_id"`
	Payload            PayloadInput   `json:"payload"`
	EventName          string         `json:"event_name"`
	EventsNames        []string       `json:"events_names"`
	Opts               Opts           `json:"opts"`
	Info               map[string]any `json:"info"`
	Type               EventType      `json:"type"`
	CreatedAt          time.Time      `json:"created_at"`
}

type PayloadInput struct {
	EventID   uuid.UUID `json:"event_id"`
	Data      []byte    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

func (p PayloadInput) Parser(v any) error {
	return json.Unmarshal(p.Data, v)
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
type ConsumerFn func(ctx context.Context, payload PayloadInput) error
