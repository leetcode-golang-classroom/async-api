package application

import (
	"context"

	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/logger"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/user"
)

func (app *App) SetupRoute(ctx context.Context) {
	slog := logger.FromContext(ctx)
	pingHandler := NewHandler(slog)
	app.router.HandleFunc("GET /ping", pingHandler.Ping)

	userStore := user.NewUserStore(app.db)
	userHandler := user.NewHandler(slog, app.validator, userStore)
	userHandler.RegisterRoute(app.router)
}
