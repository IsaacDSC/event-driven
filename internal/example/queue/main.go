package main

import (
	"context"
	"event-driven/SDK"
	"event-driven/types"
	"fmt"
	"github.com/google/uuid"
	"os"
	"os/signal"
	"syscall"
)

const serverHost = "http://localhost:3333"

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	producer := SDK.NewProducer(serverHost, nil)
	ctx := context.Background()

	if err := producer.Producer(ctx, "event_example_01", map[string]any{"key": "value"}, ConsumerExample01); err != nil {
		panic(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	consumer := SDK.NewConsumerServer("localhost:6379")

	consumer.AddHandlers(map[string]types.ConsumerFn{
		"event_example_01": ConsumerExample01,
	}).Start()

}

func ConsumerExample01(ctx context.Context, txID uuid.UUID, payload map[string]any) error {
	fmt.Printf("received_payload:: %+v\n", payload)

	//return fmt.Errorf("default error ")
	return nil
}
