package main

import (
	"context"
	"event-driven/SDK"
	"event-driven/internal/example/saga/ex"
	"event-driven/types"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	producer := SDK.NewProducer("http://localhost:3333/task", nil)

	ctx := context.Background()
	if err := producer.SagaProducer(ctx, "event_example_01", map[string]any{"key": "value"}); err != nil {
		panic(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sg1 := ex.NewSagaExample()
	sg2 := ex.NewSagaExample2()

	sp := SDK.NewSagaPattern([]SDK.ConsumerInput{
		sg1, sg2,
	}, types.Opts{
		MaxRetry: 3,
	}, false)

	consumer := SDK.NewConsumerServer("localhost:6379")

	if err := consumer.AddHandlers(map[string]types.ConsumerFn{
		"event_example_01": sp.Consumer,
	}).Start(); err != nil {
		panic(err)
	}

}
