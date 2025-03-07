# event-driven
*Library to help with distributed transaction work *

### MONITORAMENTO
TODO: [ADD IMG] 
- SERVER HTTP
    - Save transactions on database
    - Dashboard to monitor transactions
    - Removed transaction every (30 Days)

### SAGA PATTERN - ORCHESTRATOR

```go
package main

import (
	"context"
	"event-driven/SDK"
	"event-driven/internal/example/checkout/domains"
	"event-driven/types"
	"time"
)

const EventCheckoutCreated = "event.checkout.created"

func main() {
	if err := producer(); err != nil {
		panic(err)
	}

	if err := consumer(); err != nil {
		panic(err)
	}

}

func producer() error {
	pd := SDK.NewProducer("http://localhost:3333", &types.Opts{
		MaxRetry: 5,
		DeadLine: time.Now().Add(15 * time.Minute),
	})

	input := map[string]any{
		"order_id":     "79a369da-0d71-4e3f-b504-e1f793220e60",
		"client":       "John Doe",
		"client_email": "john_doe@gmail.com",
		"products": []map[string]any{
			{
				"product_id": "79a369da-0d71-4e3f-b504-e1f793220e60",
				"quantity":   1,
				"price":      100.00,
			},
			{
				"product_id": "79a369da-0d71-4e3f-b504-e1f793220e60",
				"quantity":   3,
				"price":      400.00,
			},
		},
		"total":  500.00,
		"status": "pending",
	}

	if err := pd.SagaProducer(context.Background(), EventCheckoutCreated, input); err != nil {
		return err
	}

	return nil
}

// CHECKOUT TASK EXAMPLE
func consumer() error {
	sgPayment := domains.NewPayment()
	sgStock := domains.NewStock()
	sgDelivery := domains.NewDelivery()
	sgNotify := domains.NewNotify()

	sp := SDK.NewSagaPattern([]SDK.ConsumerInput{
		sgPayment,
		sgStock,
		sgDelivery,
		sgNotify,
	}, types.Opts{
		MaxRetry: 3,
	}, false)
	consumer := SDK.NewConsumerServer("localhost:6379")

	if err := consumer.AddHandlers(map[string]types.ConsumerFn{
		EventCheckoutCreated: sp.Consumer,
	}).Start(); err != nil {
		return err
	}

	return nil
}

```


### Event Checkout Created 

```go   
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


```