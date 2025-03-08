package SDK

import (
	"context"
	"fmt"
	"github.com/IsaacDSC/event-driven/internal/mocks"
	"github.com/IsaacDSC/event-driven/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/mock/gomock"
	"testing"
)

func setupRedisContainer(t *testing.T) (testcontainers.Container, string) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis:6.2",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := redisC.Host(ctx)
	require.NoError(t, err)

	port, err := redisC.MappedPort(ctx, "6379")
	require.NoError(t, err)

	redisAddr := fmt.Sprintf("%s:%s", host, port.Port())
	return redisC, redisAddr
}

func TestConsumerServer_AddHandlers(t *testing.T) {
	redisC, redisAddr := setupRedisContainer(t)
	defer redisC.Terminate(context.Background())

	mockController := gomock.NewController(t)
	mockRepo := mocks.NewMockRepository(mockController)
	server := NewConsumerServer(redisAddr, mockRepo)
	oldMux := server.mux

	consumerFn := func(ctx context.Context, payload types.PayloadInput) error {
		return nil
	}

	consumers := map[string]types.ConsumerFn{
		"event_example": consumerFn,
	}

	server.AddHandlers(consumers)

	assert.NotNil(t, server.mux)
	assert.NotEqual(t, oldMux, server.mux)
}
