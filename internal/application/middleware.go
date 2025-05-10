package application

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/leetcode-golang-classroom/golang-async-api/internal/logger"
)

func NewLoggerMiddleware(ctx context.Context) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger.FromContext(ctx)
			log.Info("http request", slog.String("path", fmt.Sprintf("%s %s", r.Method, r.URL.Path)))
			next.ServeHTTP(w, r)
		})
	}
}
