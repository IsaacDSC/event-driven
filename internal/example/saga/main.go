package main

import (
	"context"
	"github.com/IsaacDSC/event-driven/SDK"
	"github.com/IsaacDSC/event-driven/internal/example/saga/ex"
	"github.com/IsaacDSC/event-driven/repository"
	"github.com/IsaacDSC/event-driven/types"
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

	repo, err := repository.NewPgAdapter(connectionString)
	if err != nil {
		panic(err)
	}

	defer repo.Close()

	go producerExample(repo)

	sg1 := ex.NewSagaExample()
	sg2 := ex.NewSagaExample2()

	defaultSettings := types.Opts{MaxRetry: 3}

	sp := SDK.NewSagaPattern(rdAddr, repo, []types.ConsumerInput{sg1, sg2}, defaultSettings, false)

	if err := sp.WithConsumerServer(rdAddr, repo).AddHandlers(map[string]types.ConsumerFn{
		"event_example_01": sp.Consumer,
	}).Start(); err != nil {
		panic(err)
	}

}

func producerExample(repo types.Repository) {
	producer := SDK.NewProducer(rdAddr, repo, types.EmptyOpts)

	for {
		ctx := context.Background()
		if err := producer.SagaProducer(ctx, "event_example_01", map[string]any{"key": "value"}); err != nil {
			panic(err)
		}
		time.Sleep(time.Minute * 7)
	}
}
