package SDK

import (
	"context"
	"fmt"
	"github.com/IsaacDSC/event-driven/internal/mocks"
	"github.com/IsaacDSC/event-driven/internal/utils"
	"github.com/IsaacDSC/event-driven/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestSagaPattern_Consumer(t *testing.T) {
	tests := []struct {
		name          string
		consumers     []types.ConsumerInput
		setupMocks    func(mockRepo *mocks.MockRepository, consumers []types.ConsumerInput)
		expectedError bool
	}{
		{
			name: "Successful execution",
			consumers: []types.ConsumerInput{
				mocks.NewMockConsumerInput(gomock.NewController(t)),
				mocks.NewMockConsumerInput(gomock.NewController(t)),
			},
			setupMocks: func(mockRepo *mocks.MockRepository, consumers []types.ConsumerInput) {
				mockRepo.EXPECT().SagaSaveTx(gomock.Any(), gomock.Any()).Return(nil).Times(len(consumers))
				mockRepo.EXPECT().SagaUpdateInfos(gomock.Any(), gomock.Any(), gomock.Any(), "COMMITED").Return(nil).Times(len(consumers))
				for _, consumer := range consumers {
					consumer.(*mocks.MockConsumerInput).EXPECT().UpFn(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					consumer.(*mocks.MockConsumerInput).EXPECT().GetConfig().Return(types.Opts{}).Times(1)
					consumer.(*mocks.MockConsumerInput).EXPECT().GetEventName().Return("event").AnyTimes()
				}
			},
			expectedError: false,
		},
		{
			name: "Execution with error",
			consumers: []types.ConsumerInput{
				mocks.NewMockConsumerInput(gomock.NewController(t)),
			},
			setupMocks: func(mockRepo *mocks.MockRepository, consumers []types.ConsumerInput) {
				mockRepo.EXPECT().SagaSaveTx(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				mockRepo.EXPECT().SagaUpdateInfos(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				consumers[0].(*mocks.MockConsumerInput).EXPECT().UpFn(gomock.Any(), gomock.Any()).Return(fmt.Errorf("error")).Times(2)
				consumers[0].(*mocks.MockConsumerInput).EXPECT().DownFn(gomock.Any(), gomock.Any()).Return(fmt.Errorf("error")).Times(2)
				consumers[0].(*mocks.MockConsumerInput).EXPECT().GetConfig().Return(types.Opts{MaxRetry: 1}).AnyTimes()
				consumers[0].(*mocks.MockConsumerInput).EXPECT().GetEventName().Return("event").AnyTimes()
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockController := gomock.NewController(t)
			mockRepo := mocks.NewMockRepository(mockController)
			saga := NewSagaPattern("", mockRepo, tt.consumers, types.Opts{MaxRetry: 1}, true)

			ctx := context.Background()
			payload := types.PayloadInput{
				EventID: uuid.New(),
			}
			ctx = utils.SetTxIDToCtx(ctx, uuid.New())

			tt.setupMocks(mockRepo, tt.consumers)

			err := saga.Consumer(ctx, payload)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
