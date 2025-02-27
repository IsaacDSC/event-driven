package ex

import (
	"context"
	"errors"
	"event-driven/SDK"
	"event-driven/types"
	"fmt"
	"github.com/google/uuid"
)

type SagaExample2 struct{}

var _ SDK.ConsumerInput = (*SagaExample2)(nil)

func NewSagaExample2() *SagaExample2 {
	return &SagaExample2{}
}

func (s SagaExample2) UpFn(ctx context.Context, payload types.PayloadInput) error {
	fmt.Println("UpFn Saga2 Received:", payload)
	return errors.New("generic error upfn saga2")
}

func (s SagaExample2) DownFn(ctx context.Context, payload types.PayloadInput) error {
	fmt.Println("DownFn Saga2 Received:", payload)
	return nil
}

func (s SagaExample2) GetConfig() types.Opts {
	return types.Opts{}
}

func (s SagaExample2) GetEventName() string {
	return fmt.Sprintf("%s.%s.%s", "saga", "example", uuid.New().String()[0:5])
}
