package main

import (
	"context"
	"github.com/IsaacDSC/event-driven/SDK"
	"github.com/IsaacDSC/event-driven/internal/example/checkout/domains"
	"github.com/IsaacDSC/event-driven/repository"
	"github.com/IsaacDSC/event-driven/types"
	"time"
)

const connectionString = "user=root password=root dbname=event-driven sslmode=disable"
const rdAddr = "localhost:6379"

const EventCheckoutCreated = "event.checkout.created"

func main() {
	repo, err := repository.NewPgAdapter(connectionString)
	if err != nil {
		panic(err)
	}

	defer repo.Close()

	if err := producer(repo); err != nil {
		panic(err)
	}

	if err := consumer(repo); err != nil {
		panic(err)
	}

}

func producer(repo types.Repository) error {
	pd := SDK.NewProducer(rdAddr, repo, &types.Opts{
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
func consumer(repo types.Repository) error {
	sgPayment := domains.NewPayment()
	sgStock := domains.NewStock()
	sgDelivery := domains.NewDelivery()
	sgNotify := domains.NewNotify()

	sp := SDK.NewSagaPattern(rdAddr, repo, []types.ConsumerInput{
		sgPayment,
		sgStock,
		sgDelivery,
		sgNotify,
	}, types.Opts{
		MaxRetry: 3,
	}, false)

	if err := sp.WithConsumerServer(rdAddr, repo).AddHandlers(map[string]types.ConsumerFn{
		EventCheckoutCreated: sp.Consumer,
	}).Start(); err != nil {
		return err
	}

	return nil
}
