package application

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/jwt"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/logger"
	refreshtoken "github.com/leetcode-golang-classroom/golang-async-api/internal/refresh_token"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/report"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/user"
)

func (app *App) SetupRoute(ctx context.Context) {
	slog := logger.FromContext(ctx)
	pingHandler := NewHandler(slog)
	app.router.HandleFunc("GET /ping", pingHandler.Ping)

	userStore := user.NewUserStore(app.db)
	refreshTokenStore := refreshtoken.NewRefreshTokenStore(app.db)
	jwtManager := jwt.NewJWTManager(app.config)
	app.userStore = userStore
	app.jwtManager = jwtManager
	userHandler := user.NewHandler(slog, app.validator, userStore, refreshTokenStore, jwtManager)
	userHandler.RegisterRoute(app.router)

	sdkConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load sdk config")
		os.Exit(1)
	}
	sqsClient := sqs.NewFromConfig(sdkConfig, func(options *sqs.Options) {
		options.BaseEndpoint = aws.String(app.config.LocalstackEndPoint)
	})

	s3Client := s3.NewFromConfig(sdkConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(app.config.S3LocalstackEndPoint)
		options.UsePathStyle = true
	})
	presignedClient := s3.NewPresignClient(s3Client)
	reportStore := report.NewReportStore(app.db)

	reportHandler := report.NewHandler(slog, app.validator, jwtManager,
		reportStore,
		sqsClient,
		app.config,
		presignedClient,
	)
	reportHandler.RegisterRoute(app.router)
}
