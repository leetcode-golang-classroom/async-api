package report

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/config"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/logger"
)

type ReportBuilder struct {
	appConfig    *config.Config
	resportStore *ReportStore
	lozClient    *LozClient
	s3Client     *s3.Client
}

func NewReportBuilder(
	appConfig *config.Config,
	reportStore *ReportStore,
	lozClient *LozClient,
	s3Client *s3.Client) *ReportBuilder {
	return &ReportBuilder{
		appConfig:    appConfig,
		resportStore: reportStore,
		lozClient:    lozClient,
		s3Client:     s3Client,
	}
}

func (b *ReportBuilder) Build(ctx context.Context,
	userID uuid.UUID, reportID uuid.UUID) (report *Report, err error) {
	log := logger.FromContext(ctx)
	report, err = b.resportStore.ByPrimaryKey(ctx, userID, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report %s for user %s: %w", reportID, userID, err)
	}

	if report.StartedAt.Valid {
		return report, nil
	}
	defer func() {
		if err != nil {
			now := time.Now().UTC()
			errMsg := err.Error()
			report.FailedAt = sql.NullTime{
				Time:  now,
				Valid: true,
			}
			report.ErrorMessage = sql.NullString{
				String: errMsg,
				Valid:  true,
			}
			if _, updateErr := b.resportStore.Update(ctx, report); updateErr != nil {
				log.Error("failed to update report", slog.Any("error", err))
			}
		}
	}()

	now := time.Now().UTC()
	report.StartedAt = sql.NullTime{
		Time:  now,
		Valid: true,
	}
	report, err = b.resportStore.Update(ctx, report)
	if err != nil {
		return nil, fmt.Errorf("failed to update report %s for user %d: %w", reportID, userID, err)
	}

	resp, err := b.lozClient.GetMonsters()
	if err != nil {
		return nil, fmt.Errorf("failed to get monsters data: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no monsters data found")
	}

	var buffer bytes.Buffer
	qzipWriter := gzip.NewWriter(&buffer)
	csvWriter := csv.NewWriter(qzipWriter)
	header := []string{"name", "id", "category", "description", "image", "common_locations", "drops", "dlc"}
	if err := csvWriter.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write csv header: %w", err)
	}
	for _, monster := range resp.Data {
		csvRow := []string{
			monster.Name,
			fmt.Sprintf("%d", monster.ID),
			monster.Category,
			monster.Description,
			monster.Image,
			strings.Join(monster.CommonLocations, ", "),
			strings.Join(monster.Drops, ", "),
			strconv.FormatBool(monster.Dlc),
		}
		if err := csvWriter.Write(csvRow); err != nil {
			return nil, fmt.Errorf("failed to write csv row: %w", err)
		}

		if err := csvWriter.Error(); err != nil {
			return nil, fmt.Errorf("failed to write csv row: %w", err)
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush csv writer: %w", err)
	}
	if err := qzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	key := fmt.Sprintf("/users/%s/report/%s.csv.gz", userID, reportID)
	_, err = b.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Key:    aws.String(key),
		Bucket: aws.String(b.appConfig.S3Bucket),
		Body:   bytes.NewReader(buffer.Bytes()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload report to %s: %w", key, err)
	}
	report.OutputFilePath = sql.NullString{
		String: key,
		Valid:  true,
	}
	report.CompletedAt = sql.NullTime{
		Time:  now,
		Valid: true,
	}
	report, err = b.resportStore.Update(ctx, report)
	if err != nil {
		return nil, fmt.Errorf("failed to update report %s for user %s: %w", reportID, userID, err)
	}
	log.Info("successfuly generated report", slog.String("report_id", reportID.String()),
		slog.String("user_id", userID.String()), slog.String("path", key))
	return report, nil
}
