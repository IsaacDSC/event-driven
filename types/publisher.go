package types

import (
	"context"
	"github.com/google/uuid"
)

type Producer interface {
	Producer(ctx context.Context, input PayloadType) error
	Close()
	GenerateEventID() uuid.UUID
}
