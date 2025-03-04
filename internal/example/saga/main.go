package main

import (
	"context"
	"event-driven/SDK"
	"event-driven/database"
	"event-driven/internal/example/saga/ex"
	"event-driven/repository"
	"event-driven/types"
	"log"
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

	db, err := database.NewConnection(connectionString)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	defer db.Close()

	repo := repository.New(db)

	go producerExample(repo)

	sg1 := ex.NewSagaExample()
	sg2 := ex.NewSagaExample2()

	defaultSettings := types.Opts{MaxRetry: 3}

	sp := SDK.NewSagaPattern(rdAddr, repo, []types.ConsumerInput{sg1, sg2}, defaultSettings, false)

	consumer := SDK.NewConsumerServer(rdAddr, repo)

	if err := consumer.AddHandlers(map[string]types.ConsumerFn{
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
