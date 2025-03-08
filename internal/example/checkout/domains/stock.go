package domains

import (
	"context"
	"fmt"
	"github.com/IsaacDSC/event-driven/internal/example/checkout/entities"
	"github.com/IsaacDSC/event-driven/types"
)

// SAGA TX DECREMENT PRODUCT TO STOCK
type Stock struct{}

func NewStock() *Stock {
	return &Stock{}
}

var _ types.ConsumerInput = (*Stock)(nil)

func (s Stock) UpFn(ctx context.Context, payload types.PayloadInput) error {
	var input entities.Order
	if err := payload.Parser(&input); err != nil {
		return fmt.Errorf("could not parse payload: %v", err)
	}

	stock := 100
	for _, pd := range input.Products {
		stock -= pd.Quantity
	}

	fmt.Println("UpFn Stock: ", stock)

	return nil
}

func (s Stock) DownFn(ctx context.Context, payload types.PayloadInput) error {
	var input entities.Order
	if err := payload.Parser(&input); err != nil {
		return fmt.Errorf("could not parse payload: %v", err)
	}

	stock := 100
	for _, pd := range input.Products {
		stock += pd.Quantity
	}

	fmt.Println("DownFn Stock: ", stock)

	return nil
}

func (s Stock) GetConfig() types.Opts {
	return types.Opts{
		Delay: 3,
	}
}

func (s Stock) GetEventName() string {
	return "event.stock.decremented"
}
