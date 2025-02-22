package ex

import (
	"context"
	"event-driven/SDK"
	"event-driven/types"
	"fmt"
	"github.com/google/uuid"
)

type SagaExample struct{}

var _ SDK.ConsumerInput = (*SagaExample)(nil)

func NewSagaExample() *SagaExample {
	return &SagaExample{}
}

func (s SagaExample) UpFn(ctx context.Context, payload any, opts ...types.Opts) error {
	fmt.Println("UpFn Saga1 Received:", payload)
	return nil
}

func (s SagaExample) DownFn(ctx context.Context, payload any, opts ...types.Opts) error {
	fmt.Println("DownFn Saga1 Received:", payload)
	return nil
}

func (s SagaExample) GetConfig() types.Opts {
	return types.Opts{}
}

func (s SagaExample) GetEventName() string {
	return fmt.Sprintf("%s.%s.%s", "saga", "example", uuid.New().String()[0:5])
}
