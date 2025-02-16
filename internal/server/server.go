package server

import (
	"bytes"
	"database/sql"
	"event-driven/internal/server/broker"
	"event-driven/internal/server/web"
	"event-driven/internal/sqlc"
	"event-driven/internal/utils"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"io"
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
	go broker.NewSubscriberServer("localhost:6379", s.db)

	pbServer := broker.NewProducerServer("localhost:6379")
	defer pbServer.Close()

	repository := sqlc.NewRepository(s.db)

	logs := utils.NewLogger()

	h := web.NewHandler(repository, pbServer)
	for path, fn := range h.GetRoutes() {
		http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			logs.Info("request received", "path", r.URL.Path, "method", r.Method, "body", string(b))
			r.Body = io.NopCloser(bytes.NewBuffer(b))

			if err := fn(w, r); err != nil {
				logs.Error("could not process request", "error", err)
				http.Error(w, fmt.Sprintf("could not process request: %v", err), http.StatusInternalServerError)
			}
		})
	}

	if err := http.ListenAndServe(":3333", nil); err != nil {
		return fmt.Errorf("could not start server: %v", err)
	}

	return nil
}
