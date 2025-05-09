package application

import (
	"context"

	"github.com/leetcode-golang-classroom/golang-async-api/internal/logger"
)

func (app *App) SetupRoute(ctx context.Context) {
	pingHandler := NewHandler(logger.FromContext(ctx))
	app.router.HandleFunc("GET /ping", pingHandler.Ping)
}
