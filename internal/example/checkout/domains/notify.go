package domains

import (
	"context"
	"event-driven/internal/example/checkout/entities"
	"event-driven/internal/example/checkout/fakegate"
	"event-driven/types"
	"fmt"
)

// SAGA TX SEND EMAIL TO CLIENT
type Notify struct{}

func NewNotify() *Notify {
	return &Notify{}
}

var _ types.ConsumerInput = (*Notify)(nil)

func (n Notify) UpFn(ctx context.Context, payload types.PayloadInput) error {
	var input entities.Order
	if err := payload.Parser(&input); err != nil {
		return fmt.Errorf("could not parse payload: %v", err)
	}

	err := fakegate.SentEmail(input.ClientEmail)
	if err != nil {
		return err
	}

	fmt.Println("UpFn Notify: ", input.ClientEmail)

	return nil
}

func (n Notify) DownFn(ctx context.Context, payload types.PayloadInput) error {
	return nil
}

func (n Notify) GetConfig() types.Opts {
	return types.Opts{
		Delay: 3,
	}
}

func (n Notify) GetEventName() string {
	return "event.notify.sent"
}
