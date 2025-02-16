package main

import (
	"context"
	"event-driven/internal/database"
	"event-driven/internal/server"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"log"
)

const redisChannel = "chat_channel"

func main() {
	db, err := database.NewConnection()
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	defer db.Close()

	addr := "localhost:6379"
	redisClient := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("could not connect to redis: %v", err)
	}

	s := server.NewServer(redisClient, db)
	log.Println("Server running on port 3333")
	if err := s.StartServer(); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
