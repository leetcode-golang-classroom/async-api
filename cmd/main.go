package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/leetcode-golang-classroom/golang-async-api/internal/application"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/config"
	mlog "github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/logger"
)

func main() {
	// 建立 logger
	logger := slog.New(slog.NewJSONHandler(
		os.Stdout, &slog.HandlerOptions{
			AddSource: true,
		},
	))
	rootContext := context.WithValue(context.Background(), mlog.CtxKey{}, logger)
	app := application.New(rootContext, config.AppConfig)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	err := app.Start(ctx)
	if err != nil {
		logger.Error("failed to start app", slog.Any("err", err))
	}
}
