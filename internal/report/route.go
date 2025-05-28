package report

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/config"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/helper"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/jwt"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/response"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/util"
)

type Handler struct {
	logger          *slog.Logger
	validator       *validator.Validate
	jwtManager      *jwt.JWTManager
	reportStore     *ReportStore
	sqsClient       *sqs.Client
	appConfig       *config.Config
	preSignedClient *s3.PresignClient
}

func NewHandler(logger *slog.Logger,
	validator *validator.Validate,
	jwtManager *jwt.JWTManager,
	reportStore *ReportStore,
	sqsClient *sqs.Client,
	appConfig *config.Config,
	preSignedClient *s3.PresignClient,
) *Handler {
	return &Handler{
		logger:          logger,
		validator:       validator,
		jwtManager:      jwtManager,
		reportStore:     reportStore,
		sqsClient:       sqsClient,
		appConfig:       appConfig,
		preSignedClient: preSignedClient,
	}
}

func (h *Handler) RegisterRoute(router *http.ServeMux) {
	// setup route
	router.HandleFunc("POST /reports", h.createReportHandler())
	router.HandleFunc("GET /reports/{id}", h.getReportHandler())
}

func (h *Handler) createReportHandler() http.HandlerFunc {
	return helper.Handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := helper.Decode[CreateReportRequest](r, h.validator)
		if err != nil {
			return helper.NewErrWithStatus(
				http.StatusBadRequest,
				err,
			)
		}
		defer r.Body.Close()
		user, ok := util.UserFromContext(r.Context())
		if !ok {
			return helper.NewErrWithStatus(
				http.StatusUnauthorized,
				fmt.Errorf("user not found in context"),
			)
		}

		report, err := h.reportStore.Create(r.Context(), user.ID, req.ReportType)
		if err != nil {
			return helper.NewErrWithStatus(
				http.StatusInternalServerError,
				err,
			)
		}

		sqsMessage := SQSMessage{
			UserID:   report.UserID,
			ReportID: report.ID,
		}
		bytes, err := json.Marshal(sqsMessage)
		if err != nil {
			return helper.NewErrWithStatus(
				http.StatusInternalServerError,
				err,
			)
		}
		queueURLOutput, err := h.sqsClient.GetQueueUrl(r.Context(), &sqs.GetQueueUrlInput{
			QueueName: aws.String(h.appConfig.SQSQueue),
		})
		if err != nil {
			return helper.NewErrWithStatus(
				http.StatusInternalServerError,
				err,
			)
		}
		_, err = h.sqsClient.SendMessage(r.Context(), &sqs.SendMessageInput{
			QueueUrl:    queueURLOutput.QueueUrl,
			MessageBody: aws.String(string(bytes)),
		})
		if err != nil {
			return helper.NewErrWithStatus(
				http.StatusInternalServerError,
				err,
			)
		}
		var outputFilePath *string
		if report.OutputFilePath.Valid {
			outputFilePath = &report.OutputFilePath.String
		}
		var downloadURL *string
		if report.DownloadURL.Valid {
			downloadURL = &report.DownloadURL.String
		}
		var downloadURLExpiresAt *time.Time
		if report.DownloadURLExpiresAt.Valid {
			downloadURLExpiresAt = &report.DownloadURLExpiresAt.Time
		}
		var errorMessage *string
		if report.ErrorMessage.Valid {
			errorMessage = &report.ErrorMessage.String
		}
		var startedAt *time.Time
		if report.StartedAt.Valid {
			startedAt = &report.StartedAt.Time
		}
		var completedAt *time.Time
		if report.CompletedAt.Valid {
			completedAt = &report.CompletedAt.Time
		}
		var failedAt *time.Time
		if report.FailedAt.Valid {
			failedAt = &report.FailedAt.Time
		}
		if err := helper.Encode(response.ApiResponse[ApiReport]{
			Data: &ApiReport{
				ID:                   report.ID,
				ReportType:           req.ReportType,
				OutputFilePath:       outputFilePath,
				DownloadURL:          downloadURL,
				DownloadURLExpiresAt: downloadURLExpiresAt,
				ErrorMessage:         errorMessage,
				CreatedAt:            report.CreatedAt,
				StartedAt:            startedAt,
				CompletedAt:          completedAt,
				FailedAt:             failedAt,
				Status:               report.Status(),
			},
		},
			http.StatusCreated,
			w,
		); err != nil {
			return helper.NewErrWithStatus(
				http.StatusInternalServerError,
				err,
			)
		}
		return nil
	})
}

func (h *Handler) getReportHandler() http.HandlerFunc {
	return helper.Handler(func(w http.ResponseWriter, r *http.Request) error {
		reportIDStr := r.PathValue("id")
		reportID, err := uuid.Parse(reportIDStr)
		if err != nil {
			return helper.NewErrWithStatus(
				http.StatusBadRequest,
				err,
			)
		}

		user, ok := util.UserFromContext(r.Context())
		if !ok {
			return helper.NewErrWithStatus(
				http.StatusUnauthorized,
				fmt.Errorf("user not found in context"),
			)
		}

		report, err := h.reportStore.ByPrimaryKey(r.Context(), user.ID, reportID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return helper.NewErrWithStatus(
					http.StatusNotFound,
					fmt.Errorf("report not found"),
				)
			}
			return helper.NewErrWithStatus(
				http.StatusInternalServerError,
				err,
			)
		}
		var outputFilePath *string
		if report.OutputFilePath.Valid {
			outputFilePath = &report.OutputFilePath.String
		}
		var downloadURL *string
		if report.DownloadURL.Valid {
			downloadURL = &report.DownloadURL.String
		}
		var downloadURLExpiresAt *time.Time
		if report.DownloadURLExpiresAt.Valid {
			downloadURLExpiresAt = &report.DownloadURLExpiresAt.Time
		}
		var errorMessage *string
		if report.ErrorMessage.Valid {
			errorMessage = &report.ErrorMessage.String
		}
		var startedAt *time.Time
		if report.StartedAt.Valid {
			startedAt = &report.StartedAt.Time
		}
		var completedAt *time.Time
		if report.CompletedAt.Valid {
			completedAt = &report.CompletedAt.Time
		}
		if report.CompletedAt.Valid && report.DownloadURLExpiresAt.Valid && report.DownloadURLExpiresAt.Time.Before(time.Now().UTC()) {
			completedAt = &report.CompletedAt.Time
			// to s3 client (presigned client)
			expiresAt := time.Now().Add(10 * time.Second)
			signedURL, err := h.preSignedClient.PresignGetObject(r.Context(), &s3.GetObjectInput{
				Bucket: aws.String(h.appConfig.S3Bucket),
				Key:    aws.String(report.OutputFilePath.String),
			}, func(options *s3.PresignOptions) {
				options.Expires = time.Second * 10
			})
			if err != nil {
				return helper.NewErrWithStatus(
					http.StatusInternalServerError,
					err,
				)
			}
			report.DownloadURL = sql.NullString{
				String: signedURL.URL,
				Valid:  true,
			}
			report.DownloadURLExpiresAt = sql.NullTime{
				Time:  expiresAt,
				Valid: true,
			}
			report, err = h.reportStore.Update(r.Context(), report)
			if err != nil {
				return helper.NewErrWithStatus(
					http.StatusInternalServerError,
					err,
				)
			}
		}
		var failedAt *time.Time
		if report.FailedAt.Valid {
			failedAt = &report.FailedAt.Time
		}
		if err := helper.Encode(response.ApiResponse[ApiReport]{
			Data: &ApiReport{
				ID:                   report.ID,
				ReportType:           report.ReportType,
				OutputFilePath:       outputFilePath,
				DownloadURL:          downloadURL,
				DownloadURLExpiresAt: downloadURLExpiresAt,
				ErrorMessage:         errorMessage,
				CreatedAt:            report.CreatedAt,
				StartedAt:            startedAt,
				CompletedAt:          completedAt,
				FailedAt:             failedAt,
				Status:               report.Status(),
			},
		},
			http.StatusOK,
			w,
		); err != nil {
			return helper.NewErrWithStatus(
				http.StatusInternalServerError,
				err,
			)
		}
		return nil
	})
}
