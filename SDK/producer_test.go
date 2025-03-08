package SDK

import (
	"context"
	"github.com/IsaacDSC/event-driven/internal/mocks"
	"github.com/IsaacDSC/event-driven/types"
	"go.uber.org/mock/gomock"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProducer_SagaProducer(t *testing.T) {
	mockController := gomock.NewController(t)

	mockRepo := mocks.NewMockRepository(mockController)
	mockPublisher := mocks.NewMockProducer(mockController)
	mockDateTime := mocks.NewMockTimeProvider(mockController)
	producer := &Producer{
		repository: mockRepo,
		pb:         mockPublisher,
		defaultOpts: &types.Opts{
			MaxRetry: 10,
		},
		dt: mockDateTime,
	}
	ctx := context.Background()
	eventName := "test_event"
	payload := map[string]any{"key": "value"}

	inputPayload := types.PayloadInput{
		EventID:   uuid.New(),
		CreatedAt: time.Now(),
		Data:      []byte(`{"key":"value"}`),
	}

	input := types.PayloadType{
		EventID:   inputPayload.EventID,
		Payload:   inputPayload,
		EventName: eventName,
		CreatedAt: inputPayload.CreatedAt,
		Type:      types.EventTypeSaga,
		Opts:      *producer.defaultOpts,
	}

	mockPublisher.EXPECT().GenerateEventID().Return(input.EventID).Times(2)
	mockDateTime.EXPECT().Now().Return(input.CreatedAt).Times(2)
	mockRepo.EXPECT().SaveTx(gomock.Any(), input).Return(nil)
	mockPublisher.EXPECT().Producer(gomock.Any(), input).Return(nil)

	err := producer.SagaProducer(ctx, eventName, payload)
	assert.NoError(t, err)
}

func TestProducer_Producer(t *testing.T) {
	mockController := gomock.NewController(t)

	mockRepo := mocks.NewMockRepository(mockController)
	mockPublisher := mocks.NewMockProducer(mockController)
	mockDateTime := mocks.NewMockTimeProvider(mockController)
	producer := &Producer{
		repository: mockRepo,
		pb:         mockPublisher,
		defaultOpts: &types.Opts{
			MaxRetry: 10,
		},
		dt: mockDateTime,
	}

	ctx := context.Background()
	eventName := "test_event"
	payload := map[string]any{"key": "value"}

	inputPayload := types.PayloadInput{
		EventID:   uuid.New(),
		CreatedAt: time.Now(),
		Data:      []byte(`{"key":"value"}`),
	}

	input := types.PayloadType{
		EventID:   inputPayload.EventID,
		Payload:   inputPayload,
		EventName: eventName,
		CreatedAt: inputPayload.CreatedAt,
		Type:      types.EventTypeTask,
		Opts:      *producer.defaultOpts,
	}

	mockPublisher.EXPECT().GenerateEventID().Return(input.EventID).Times(2)
	mockDateTime.EXPECT().Now().Return(input.CreatedAt).Times(2)
	mockRepo.EXPECT().SaveTx(gomock.Any(), input).Return(nil)
	mockPublisher.EXPECT().Producer(gomock.Any(), input).Return(nil)

	err := producer.Producer(ctx, eventName, payload, nil)
	assert.NoError(t, err)
}
