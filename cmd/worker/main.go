package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/config"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/db"
	mlog "github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/logger"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/report"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(
		os.Stdout, &slog.HandlerOptions{
			AddSource: true,
		},
	))
	rootContext := context.WithValue(context.Background(), mlog.CtxKey{}, logger)
	ctx, cancel := signal.NotifyContext(rootContext, os.Interrupt)
	defer cancel()

	appConfig := config.AppConfig

	rdb, err := db.Connect(appConfig.DBURL)
	if err != nil {
		return err
	}
	reportStore := report.NewReportStore(rdb)
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	s3Client := s3.NewFromConfig(awsConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(appConfig.S3LocalstackEndPoint)
		options.UsePathStyle = true
	})

	sqsClient := sqs.NewFromConfig(awsConfig, func(options *sqs.Options) {
		options.BaseEndpoint = aws.String(appConfig.LocalstackEndPoint)
	})

	lozClient := report.NewClient(&http.Client{
		Timeout: 10 * time.Second,
	})

	builder := report.NewReportBuilder(appConfig, reportStore, lozClient, s3Client)

	maxConcurrency := 2
	worker := report.NewWorker(appConfig, builder, logger, sqsClient, int32(maxConcurrency))
	if err := worker.Start(ctx); err != nil {
		return err
	}
	return nil
}
