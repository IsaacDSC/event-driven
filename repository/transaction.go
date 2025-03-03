package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	genrepo "event-driven/internal/sqlc/generated/repository"
	"event-driven/types"
	"fmt"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
	"time"
)

type Transaction struct {
	orm *genrepo.Queries
}

// Ensure Transaction implements Repository
var _ types.Repository = (*Transaction)(nil)

func NewTransaction(orm *genrepo.Queries) *Transaction {
	return &Transaction{orm: orm}
}

func (c Transaction) UpdateInfos(ctx context.Context, ID uuid.UUID, retry int, status string) error {
	//TODO: review this responsibility
	var finishedAt time.Time
	if status == "FINISHED" || status == "BACKWARD" || status == "BACKWARD_ERROR" {
		finishedAt = time.Now()
	}

	input := types.UpdatePayloadInput{
		Status:     status,
		TotalRetry: retry,
		FinishedAt: finishedAt,
	}

	if err := c.orm.UpdateTransaction(ctx, genrepo.UpdateTransactionParams{
		EventID: ID,
		Status:  input.Status,
		TotalRetry: sql.NullInt32{
			Int32: int32(input.TotalRetry),
			Valid: true,
		},
		EndedAt: sql.NullTime{
			Time:  input.FinishedAt,
			Valid: true,
		},
	}); err != nil {
		return fmt.Errorf("could not update transaction with error: %v\n", err)
	}

	return nil
}

func (c Transaction) SaveTx(ctx context.Context, input types.PayloadType) error {
	opts, _ := json.Marshal(input.Opts)
	payload, _ := json.Marshal(input.Payload)
	info, _ := json.Marshal(input.Info)

	if err := c.orm.CreateTransaction(ctx, genrepo.CreateTransactionParams{
		EventID:   input.EventID,
		EventName: input.EventName,
		Opts:      opts,
		StartedAt: sql.NullTime{
			Time:  input.CreatedAt,
			Valid: true,
		},
		Info: pqtype.NullRawMessage{
			RawMessage: info,
			Valid:      true,
		},
		Payload: payload,
		Status:  "PENDING",
	}); err != nil {
		return fmt.Errorf("could not create transaction with error: %v\n", err)
	}

	return nil
}
