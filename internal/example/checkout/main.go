package main

import (
	"context"
	"event-driven/SDK"
	"event-driven/database"
	"event-driven/internal/example/checkout/domains"
	"event-driven/repository"
	"event-driven/types"
	"log"
	"time"
)

const connectionString = "user=root password=root dbname=event-driven sslmode=disable"
const rdAddr = "localhost:6379"

const EventCheckoutCreated = "event.checkout.created"

func main() {
	db, err := database.NewConnection(connectionString)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	defer db.Close()

	repo := repository.New(db)

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

	//TODO: adicionar uma espécie de herança de saga e consumer
	sp := SDK.NewSagaPattern(rdAddr, repo, []SDK.ConsumerInput{
		sgPayment,
		sgStock,
		sgDelivery,
		sgNotify,
	}, types.Opts{
		MaxRetry: 3,
	}, false)

	consumer := SDK.NewConsumerServer(rdAddr, repo)

	if err := consumer.AddHandlers(map[string]types.ConsumerFn{
		EventCheckoutCreated: sp.Consumer,
	}).Start(); err != nil {
		return err
	}

	return nil
}
