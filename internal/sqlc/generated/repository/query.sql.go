// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package genrepo

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

const createTransaction = `-- name: CreateTransaction :exec
INSERT INTO transactions (id,event_id, event_name, opts, payload, status, started_at, info)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

type CreateTransactionParams struct {
	ID        uuid.UUID
	EventID   uuid.UUID
	EventName string
	Opts      json.RawMessage
	Payload   json.RawMessage
	Status    string
	StartedAt sql.NullTime
	Info      pqtype.NullRawMessage
}

func (q *Queries) CreateTransaction(ctx context.Context, arg CreateTransactionParams) error {
	_, err := q.db.ExecContext(ctx, createTransaction,
		arg.ID,
		arg.EventID,
		arg.EventName,
		arg.Opts,
		arg.Payload,
		arg.Status,
		arg.StartedAt,
		arg.Info,
	)
	return err
}

const createTxSaga = `-- name: CreateTxSaga :exec
INSERT INTO tx_sagas (event_id, transaction_id, event_name, opts, payload, status, started_at, info, total_retry,
                      retries_errors)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT (event_id) DO
UPDATE SET status = EXCLUDED.status,
    total_retry = EXCLUDED.total_retry,
    retries_errors = EXCLUDED.retries_errors,
    updated_at = NOW()
`

type CreateTxSagaParams struct {
	EventID       uuid.UUID
	TransactionID uuid.UUID
	EventName     string
	Opts          json.RawMessage
	Payload       json.RawMessage
	Status        string
	StartedAt     sql.NullTime
	Info          pqtype.NullRawMessage
	TotalRetry    sql.NullInt32
	RetriesErrors pqtype.NullRawMessage
}

func (q *Queries) CreateTxSaga(ctx context.Context, arg CreateTxSagaParams) error {
	_, err := q.db.ExecContext(ctx, createTxSaga,
		arg.EventID,
		arg.TransactionID,
		arg.EventName,
		arg.Opts,
		arg.Payload,
		arg.Status,
		arg.StartedAt,
		arg.Info,
		arg.TotalRetry,
		arg.RetriesErrors,
	)
	return err
}

const updateTransaction = `-- name: UpdateTransaction :exec
UPDATE transactions
SET status      = $2,
    total_retry = $3,
    updated_at  = NOW(),
    ended_at    = $4
WHERE event_id = $1
`

type UpdateTransactionParams struct {
	EventID    uuid.UUID
	Status     string
	TotalRetry sql.NullInt32
	EndedAt    sql.NullTime
}

func (q *Queries) UpdateTransaction(ctx context.Context, arg UpdateTransactionParams) error {
	_, err := q.db.ExecContext(ctx, updateTransaction,
		arg.EventID,
		arg.Status,
		arg.TotalRetry,
		arg.EndedAt,
	)
	return err
}

const updateTxSaga = `-- name: UpdateTxSaga :exec
UPDATE tx_sagas
SET status         = $2,
    total_retry    = $3,
    updated_at     = NOW(),
    ended_at       = $4,
    retries_errors = $5
WHERE event_id = $1
`

type UpdateTxSagaParams struct {
	EventID       uuid.UUID
	Status        string
	TotalRetry    sql.NullInt32
	EndedAt       sql.NullTime
	RetriesErrors pqtype.NullRawMessage
}

func (q *Queries) UpdateTxSaga(ctx context.Context, arg UpdateTxSagaParams) error {
	_, err := q.db.ExecContext(ctx, updateTxSaga,
		arg.EventID,
		arg.Status,
		arg.TotalRetry,
		arg.EndedAt,
		arg.RetriesErrors,
	)
	return err
}
