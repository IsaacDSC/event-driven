package types

import (
	"context"
	"github.com/google/uuid"
)

// Repository defines the methods that the Client struct must implement
type Repository interface {
	UpdateInfos(ctx context.Context, txID uuid.UUID, retry int, status string) error
	SaveTx(ctx context.Context, input PayloadType) error
	SagaUpdateInfos(ctx context.Context, txID uuid.UUID, retry int, status string) error
	SagaSaveTx(ctx context.Context, input PayloadType) error
	Close()
}
