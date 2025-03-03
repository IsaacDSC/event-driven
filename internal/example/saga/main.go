package main

import (
	"context"
	"event-driven/SDK"
	"event-driven/internal/example/saga/ex"
	"event-driven/types"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const connectionString = "user=root password=root dbname=event-driven sslmode=disable"
const rdAddr = "localhost:6379"

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go producerExample()

	sg1 := ex.NewSagaExample()
	sg2 := ex.NewSagaExample2()

	conn := types.Connection{
		Database:  connectionString,
		RedisAddr: rdAddr,
	}

	defaultSettings := types.Opts{MaxRetry: 3}

	sp := SDK.NewSagaPattern(conn, []SDK.ConsumerInput{sg1, sg2}, defaultSettings, false)

	consumer := SDK.NewConsumerServer(conn)

	if err := consumer.AddHandlers(map[string]types.ConsumerFn{
		"event_example_01": sp.Consumer,
	}).Start(); err != nil {
		panic(err)
	}

}

func producerExample() {
	producer := SDK.NewProducer(types.Connection{
		Database:  connectionString,
		RedisAddr: rdAddr,
	}, types.EmptyOpts)

	for {
		ctx := context.Background()
		if err := producer.SagaProducer(ctx, "event_example_01", map[string]any{"key": "value"}); err != nil {
			panic(err)
		}
		time.Sleep(time.Minute * 7)
	}
}
