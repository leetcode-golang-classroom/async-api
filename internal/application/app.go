package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/config"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/logger"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/util"
)

// App - for server dependency
type App struct {
	config *config.Config
	router *http.ServeMux
}

func New(ctx context.Context, config *config.Config) *App {
	app := &App{
		config: config,
		router: http.NewServeMux(),
	}
	app.SetupRoute(ctx)
	return app
}

func (app *App) Start(ctx context.Context) error {
	middleware := NewLoggerMiddleware(ctx)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", app.config.Port),
		Handler: middleware(app.router),
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
