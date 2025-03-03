package server

import (
	"database/sql"
	broker2 "event-driven/broker"
	"event-driven/internal/server/web"
	"event-driven/internal/sqlc"
	"event-driven/internal/utils"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"net/http"
)

type Server struct {
	client *redis.Client
	db     *sql.DB
}

func NewServer(client *redis.Client, db *sql.DB) *Server {
	return &Server{client: client, db: db}
}

func (s *Server) StartServer() error {
	go broker2.NewSubscriberServer("localhost:6379", s.db)

	pbServer := broker2.NewProducerServer("localhost:6379")
	defer pbServer.Close()

	repository := sqlc.NewRepository(s.db)

	logs := utils.NewLogger("[*] ")

	h := web.NewHandler(repository, pbServer)
	for path, fn := range h.GetRoutes() {
		http.HandleFunc(path, web.LoggingMiddleware(logs, func(w http.ResponseWriter, r *http.Request) {
			if err := fn(w, r); err != nil {
				fmt.Println("retrieve error", err.Error())
				http.Error(w, fmt.Sprintf("could not process request: %v", err), http.StatusInternalServerError)
			}
		}))
	}

	if err := http.ListenAndServe(":3333", nil); err != nil {
		return fmt.Errorf("could not start server: %v", err)
	}

	return nil
}
