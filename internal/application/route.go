package application

import (
	"context"

	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/jwt"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/logger"
	refreshtoken "github.com/leetcode-golang-classroom/golang-async-api/internal/refresh_token"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/user"
)

func (app *App) SetupRoute(ctx context.Context) {
	slog := logger.FromContext(ctx)
	pingHandler := NewHandler(slog)
	app.router.HandleFunc("GET /ping", pingHandler.Ping)

	userStore := user.NewUserStore(app.db)
	refreshTokenStore := refreshtoken.NewRefreshTokenStore(app.db)
	jwtManager := jwt.NewJWTManager(app.config)
	userHandler := user.NewHandler(slog, app.validator, userStore, refreshTokenStore, jwtManager)
	userHandler.RegisterRoute(app.router)
}
