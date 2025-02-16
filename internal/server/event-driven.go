package server

import (
	"database/sql"
	"event-driven/SDK"
	"event-driven/internal/server/broker"
)

type EventDriven struct {
	producer *broker.PublisherServer
	consumer *SDK.ConsumerServer
}

func NewEventDriven(addr string, db *sql.DB) *EventDriven {
	producer := broker.NewProducerServer(addr)
	consumer := SDK.NewConsumerServer(addr)
	return &EventDriven{
		producer: producer,
		consumer: consumer,
	}
}
