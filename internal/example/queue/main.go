package main

import (
	"context"
	"event-driven/SDK"
	"event-driven/types"
	"fmt"

	"os"
	"os/signal"
	"syscall"
)

const connectionString = "user=root password=root dbname=event-driven sslmode=disable"
const rdAddr = "localhost:6379"

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	conn := types.Connection{
		Database:  connectionString,
		RedisAddr: rdAddr,
	}

	producer := SDK.NewProducer(conn, types.EmptyOpts)

	defer producer.Close()

	ctx := context.Background()
	if err := producer.Producer(ctx, "event_example_01", map[string]any{"key": "value"}, ConsumerExample01); err != nil {
		panic(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	consumer := SDK.NewConsumerServer(conn)

	if err := consumer.AddHandlers(map[string]types.ConsumerFn{
		"event_example_01": ConsumerExample01,
	}).Start(); err != nil {
		panic(err)
	}

}

func ConsumerExample01(ctx context.Context, payload types.PayloadInput) error {
	fmt.Printf("received_payload:: %+v\n", payload)
	v := make(map[string]any)
	if err := payload.Parser(&v); err != nil {
		return err
	}

	fmt.Println(v)

	return nil
}
