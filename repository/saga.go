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

type Saga struct {
	orm *genrepo.Queries
}

var _ types.Repository = (*Saga)(nil)

func NewSaga(orm *genrepo.Queries) *Saga {
	return &Saga{orm: orm}
}

func (s Saga) UpdateInfos(ctx context.Context, ID uuid.UUID, retry int, status string) error {
	if err := s.orm.UpdateTxSaga(ctx, genrepo.UpdateTxSagaParams{
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

func (s Saga) SaveTx(ctx context.Context, input types.PayloadType) error {
	opts, _ := json.Marshal(input.Opts)
	payload, _ := json.Marshal(input.Payload)
	info, _ := json.Marshal(input.Info)

	transaction, err := s.orm.GetTransactionByEventID(ctx, input.TransactionEventID)
	if err != nil {
		return fmt.Errorf("could not get transaction with error: %v\n", err)
	}

	if err := s.orm.CreateTxSaga(ctx, genrepo.CreateTxSagaParams{
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
