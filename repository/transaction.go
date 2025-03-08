package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/IsaacDSC/event-driven/database"
	genrepo "github.com/IsaacDSC/event-driven/internal/sqlc/generated/repository"
	"github.com/IsaacDSC/event-driven/types"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
	"time"
)

type Transaction struct {
	db  *sql.DB
	orm *genrepo.Queries
}

// Ensure Transaction implements Repository
var _ types.Repository = (*Transaction)(nil)

func NewPgAdapter(connStr string) (*Transaction, error) {
	r := new(Transaction)
	db, err := database.NewConnection(connStr)
	if err != nil {
		return r, fmt.Errorf("could not connect to database: %w", err)
	}

	r.orm = genrepo.New(db)

	return r, nil
}

func (t Transaction) UpdateInfos(ctx context.Context, ID uuid.UUID, retry int, status string) error {
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

	if err := t.orm.UpdateTransaction(ctx, genrepo.UpdateTransactionParams{
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

func (t Transaction) SaveTx(ctx context.Context, input types.PayloadType) error {
	opts, _ := json.Marshal(input.Opts)
	payload, _ := json.Marshal(input.Payload)
	info, _ := json.Marshal(input.Info)

	if err := t.orm.CreateTransaction(ctx, genrepo.CreateTransactionParams{
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

func (t Transaction) SagaUpdateInfos(ctx context.Context, ID uuid.UUID, retry int, status string) error {
	if err := t.orm.UpdateTxSaga(ctx, genrepo.UpdateTxSagaParams{
		EventID: ID,
		Status:  status,
		TotalRetry: sql.NullInt32{
			Int32: int32(retry),
			Valid: true,
		},
		EndedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	}); err != nil {
		return fmt.Errorf("could not update transaction with error: %v\n", err)
	}

	return nil
}

func (t Transaction) SagaSaveTx(ctx context.Context, input types.PayloadType) error {
	opts, _ := json.Marshal(input.Opts)
	payload, _ := json.Marshal(input.Payload)
	info, _ := json.Marshal(input.Info)

	transaction, err := t.orm.GetTransactionByEventID(ctx, input.TransactionEventID)
	if err != nil {
		return fmt.Errorf("could not get transaction with error: %v\n", err)
	}

	if err := t.orm.CreateTxSaga(ctx, genrepo.CreateTxSagaParams{
		TransactionID: transaction.ID,
		EventID:       input.EventID,
		EventName:     input.EventName,
		Opts:          opts,
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
		return fmt.Errorf("could not create tx_sagas with error: %v\n", err)
	}

	return nil
}

func (t Transaction) Close() {
	t.db.Close()
}
