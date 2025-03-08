package repository

import (
	"context"
	"database/sql"
	"fmt"
	genrepo "github.com/IsaacDSC/event-driven/internal/sqlc/generated/repository"
	"github.com/IsaacDSC/event-driven/types"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupPostgresContainer(t *testing.T) (testcontainers.Container, *sql.DB) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(2 * time.Minute),
	}
	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := postgresC.Host(ctx)
	require.NoError(t, err)

	port, err := postgresC.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("host=%s port=%s user=postgres password=password dbname=testdb sslmode=disable", host, port.Port())
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	// Execute the SQL schema
	schemaFile := "../internal/sqlc/schema.sql"
	schema, err := os.ReadFile(schemaFile)
	require.NoError(t, err)
	_, err = db.Exec(string(schema))
	require.NoError(t, err)

	return postgresC, db
}

func TestTransaction_UpdateInfos(t *testing.T) {
	postgresC, db := setupPostgresContainer(t)
	defer postgresC.Terminate(context.Background())
	defer db.Close()

	// Initialize the Transaction repository
	repo := Transaction{
		orm: genrepo.New(db),
	}

	// Create a sample transaction
	ctx := context.Background()
	eventID := uuid.New()
	input := types.PayloadType{
		EventID:   eventID,
		EventName: "test_event",
		CreatedAt: time.Now(),
	}
	err := repo.SaveTx(ctx, input)
	require.NoError(t, err)

	// Update the transaction
	err = repo.UpdateInfos(ctx, eventID, 1, "FINISHED")
	require.NoError(t, err)

	// Verify the update
	transaction, err := repo.orm.GetTransactionByEventID(ctx, eventID)
	require.NoError(t, err)
	assert.Equal(t, "FINISHED", transaction.Status)
	assert.Equal(t, int32(1), transaction.TotalRetry.Int32)
	assert.True(t, transaction.EndedAt.Valid)
}

func TestTransaction_SaveTx(t *testing.T) {
	postgresC, db := setupPostgresContainer(t)
	defer postgresC.Terminate(context.Background())
	defer db.Close()

	// Initialize the Transaction repository
	repo := Transaction{
		orm: genrepo.New(db),
	}

	// Create a sample transaction
	ctx := context.Background()
	eventID := uuid.New()
	input := types.PayloadType{
		EventID:   eventID,
		EventName: "test_event",
		CreatedAt: time.Now(),
	}
	err := repo.SaveTx(ctx, input)
	require.NoError(t, err)

	// Verify the transaction was saved
	transaction, err := repo.orm.GetTransactionByEventID(ctx, eventID)
	require.NoError(t, err)
	assert.Equal(t, "test_event", transaction.EventName)
	assert.Equal(t, "PENDING", transaction.Status)
}
