package application

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/config"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/db"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/jwt"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/logger"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/util"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/user"
)

// App - for server dependency
type App struct {
	config     *config.Config
	router     *http.ServeMux
	validator  *validator.Validate
	db         *sql.DB
	userStore  *user.UserStore
	jwtManager *jwt.JWTManager
}

func New(ctx context.Context, config *config.Config) *App {
	db, err := db.Connect(config.DBURL)
	log := logger.FromContext(ctx)
	if err != nil {
		log.ErrorContext(ctx, "failed to connect to db", slog.Any("err", err))
		os.Exit(1)
	}
	app := &App{
		config:    config,
		router:    http.NewServeMux(),
		validator: validator.New(validator.WithRequiredStructEnabled()),
		db:        db,
	}
	app.SetupRoute(ctx)
	return app
}

func (app *App) Start(ctx context.Context) error {
	middleware := NewLoggerMiddleware(ctx)
	authMiddleware := NewAuthMiddleware(ctx, app.jwtManager, app.userStore)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", app.config.Port),
		Handler: authMiddleware(middleware(app.router)),
	}
	log := logger.FromContext(ctx)
	log.Info(fmt.Sprintf("starting server on %s", app.config.Port))
	var err error
	errCh := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			errCh <- fmt.Errorf("failed to start server: %w", err)
		}
		util.CloseChannel(errCh)
	}()
	select {
	case err = <-errCh:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		log.Warn("stopping server, wait for 10 seconds to stop")
		defer cancel()
		return server.Shutdown(timeout)
	}
}
