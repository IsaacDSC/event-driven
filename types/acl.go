package types

import (
	"context"
	"github.com/google/uuid"
)

// ClientInterface defines the methods that the Client struct must implement
type ClientInterface interface {
	UpdateInfos(ctx context.Context, txID uuid.UUID, retry int, status string)
	CreateMsg(ctx context.Context, input PayloadType) error
}
