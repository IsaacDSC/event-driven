package web

import (
	"event-driven/internal/utils"
	"log/slog"
	"net/http"
)

func LoggingMiddleware(logger *utils.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Request received", slog.String("method", r.Method), slog.String("path", r.URL.Path))
		next.ServeHTTP(w, r)
		logger.Info("Request processed")
	}
}
