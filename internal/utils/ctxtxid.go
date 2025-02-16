package utils

import (
	"context"
	"errors"
	"github.com/google/uuid"
)

var txIDKey struct{}

func SetTxIDToCtx(ctx context.Context, value any) context.Context {
	return context.WithValue(ctx, txIDKey, value)
}

func GetTxIDFromCtx(ctx context.Context) (uuid.UUID, error) {
	switch ctx.Value(txIDKey).(type) {
	case uuid.UUID:
		return ctx.Value(txIDKey).(uuid.UUID), nil
	case string:
		txID, err := uuid.Parse(ctx.Value(txIDKey).(string))
		return txID, err
	default:
		return uuid.Nil, errors.New("could not parse txID from context")
	}
}
