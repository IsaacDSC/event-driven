package ex

import (
	"context"
	"fmt"
	"github.com/IsaacDSC/event-driven/types"
)

type SagaExample struct{}

var _ types.ConsumerInput = (*SagaExample)(nil)

func NewSagaExample() *SagaExample {
	return &SagaExample{}
}

func (s SagaExample) UpFn(ctx context.Context, payload types.PayloadInput) error {
	fmt.Println("UpFn Saga1 Received:", payload)
	return nil
}

func (s SagaExample) DownFn(ctx context.Context, payload types.PayloadInput) error {
	fmt.Println("DownFn Saga1 Received:", payload)
	return nil
}

func (s SagaExample) GetConfig() types.Opts {
	return types.Opts{}
}

func (s SagaExample) GetEventName() string {
	return fmt.Sprintf("%s.%s.%s", "saga", "example", "1")
}
