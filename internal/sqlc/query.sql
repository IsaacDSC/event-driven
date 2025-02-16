-- name: CreateTransaction :exec
INSERT INTO transactions (event_id, event_name, opts, payload, status, started_at, info)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: UpdateTransaction :exec
UPDATE transactions
SET status      = $2,
    total_retry = $3,
    updated_at  = NOW(),
    ended_at    = $4
WHERE event_id = $1;


-- name: CreateTxSaga :exec
INSERT INTO tx_sagas (event_id, transaction_id, event_name, opts, payload, status, started_at, info, total_retry,
                      retries_errors)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (event_id) DO UPDATE SET status         = EXCLUDED.status,
                                     total_retry    = EXCLUDED.total_retry,
                                     retries_errors = EXCLUDED.retries_errors,
                                     updated_at     = NOW();


-- name: UpdateTxSaga :exec
UPDATE tx_sagas
SET status         = $2,
    total_retry    = $3,
    updated_at     = NOW(),
    ended_at       = $4,
    retries_errors = $5
WHERE event_id = $1;


-- name: GetTransactionByEventID :one
SELECT * FROM transactions WHERE event_id = $1;