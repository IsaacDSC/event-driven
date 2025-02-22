package domains

import (
	"context"
	"event-driven/SDK"
	"event-driven/internal/example/checkout/entities"
	"event-driven/internal/example/checkout/fakegate"
	"event-driven/types"
	"fmt"
)

// SAGA TX scheduler Payment
type Payment struct{}

func NewPayment() *Payment {
	return &Payment{}
}

var _ SDK.ConsumerInput = (*Payment)(nil)

func (p Payment) UpFn(ctx context.Context, payload types.PayloadInput) error {
	var input entities.Order
	if err := payload.Parser(&input); err != nil {
		return fmt.Errorf("could not parse payload: %v", err)
	}

	if err := fakegate.SentPayment(input.ClientEmail, input.Total); err != nil {
		return err
	}

	fmt.Println("UpFn Payment: ", input.Products)

	return nil
}

func (p Payment) DownFn(ctx context.Context, payload types.PayloadInput) error {
	return nil
}

func (p Payment) GetConfig() types.Opts {
	return types.Opts{
		Delay: 3,
	}
}

func (p Payment) GetEventName() string {
	return "event.payment.charged"
}
