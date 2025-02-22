package domains

import (
	"context"
	"event-driven/SDK"
	"event-driven/internal/example/checkout/entities"
	"event-driven/internal/example/checkout/fakegate"
	"event-driven/types"
	"fmt"
)

// SAGA TX scheduler Delivery
type Delivery struct{}

func NewDelivery() *Delivery {
	return &Delivery{}
}

var _ SDK.ConsumerInput = (*Delivery)(nil)

func (d Delivery) UpFn(ctx context.Context, payload types.PayloadInput) error {
	var input entities.Order
	if err := payload.Parser(&input); err != nil {
		return fmt.Errorf("could not parse payload: %v", err)
	}

	if err := fakegate.SentToScheduleDelivery(input.Products); err != nil {
		return fmt.Errorf("failed to schedule delivery: %v", err)
	}

	fmt.Println("UpFn Delivery: ", input.Products)

	return nil
}

func (d Delivery) DownFn(ctx context.Context, payload types.PayloadInput) error {
	var input entities.Order
	if err := payload.Parser(input); err != nil {
		return fmt.Errorf("could not parse payload: %v", err)
	}

	if err := fakegate.SentToScheduleDelivery(input.Products); err != nil {
		return fmt.Errorf("failed to schedule delivery: %v", err)
	}

	fmt.Println("DownFn Revert Delivery: ", input.Products)

	return nil
}

func (d Delivery) GetConfig() types.Opts {
	return types.Opts{
		Delay: 3,
	}
}

func (d Delivery) GetEventName() string {
	return "event.delivery.scheduled"
}
