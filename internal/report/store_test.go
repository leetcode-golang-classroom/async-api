package report_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/config"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/db"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/report"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func SetupDB(t *testing.T) (*sql.DB, *migrate.Migrate) {
	t.Helper()
	appConfig := config.AppConfig
	appConfig.SetupEnv(config.Env_Dev)
	dbURL := appConfig.DBURLTEST
	db, err := db.Connect(dbURL)
	require.NoError(t, err)

	result := strings.Replace(appConfig.PROJECT_ROOT, "/internal/report", "", 1)
	m, err := migrate.New(
		fmt.Sprintf("file://%s/migrations", result),
		dbURL,
	)
	require.NoError(t, err)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}
	return db, m
}

func TestReportStore(t *testing.T) {
	db, m := SetupDB(t)
	defer db.Close()
	ctx := context.Background()
	reportStore := report.NewReportStore(db)
	userStore := user.NewUserStore(db)
	user1, err := userStore.CreateUser(ctx, "test@test.com", "secretpassword")
	require.NoError(t, err)

	now := time.Now().UTC()
	report1, err := reportStore.Create(ctx, user1.ID, "monsters")
	require.NoError(t, err)
	assert.Equal(t, user1.ID, report1.UserID)
	assert.Equal(t, "monsters", report1.ReportType)
	assert.Less(t, now.UnixNano(), report1.CreatedAt.UnixNano())
	startedAt := report1.CreatedAt.Add(time.Second)
	completedAt := report1.CreatedAt.Add(2 * time.Second)
	failedAt := report1.CreatedAt.Add(2 * time.Second)
	report1.ReportType = "food"
	report1.StartedAt = sql.NullTime{
		Time:  startedAt,
		Valid: true,
	}
	report1.CompletedAt = sql.NullTime{
		Time:  completedAt,
		Valid: true,
	}
	report1.FailedAt = sql.NullTime{
		Time:  failedAt,
		Valid: true,
	}
	errorMsg := "there was a failure"
	report1.ErrorMessage = sql.NullString{
		String: errorMsg,
		Valid:  true,
	}
	downloadURL := "http://localhost:8080"
	report1.DownloadURL = sql.NullString{
		String: downloadURL,
		Valid:  true,
	}
	outputPath := "s3://reports-test/reports"
	report1.OutputFilePath = sql.NullString{
		String: outputPath,
		Valid:  true,
	}
	downloadURLExpiresAt := report1.CreatedAt.Add(4 * time.Second)
	report1.DownloadURLExpiresAt = sql.NullTime{
		Time:  downloadURLExpiresAt,
		Valid: true,
	}
	report2, err := reportStore.Update(ctx, report1)
	require.NoError(t, err)
	assert.Equal(t, report1.ID, report2.ID)
	assert.Equal(t, report1.CreatedAt.UnixNano(), report2.CreatedAt.UnixNano())
	assert.Equal(t, report1.StartedAt.Time.UnixNano(), report2.StartedAt.Time.UnixNano())
	assert.Equal(t, report1.CompletedAt.Time.UnixNano(), report2.CompletedAt.Time.UnixNano())
	assert.Equal(t, report1.FailedAt.Time.UnixNano(), report2.FailedAt.Time.UnixNano())
	assert.Equal(t, errorMsg, report2.ErrorMessage.String)
	assert.Equal(t, downloadURL, report2.DownloadURL.String)
	assert.Equal(t, outputPath, report2.OutputFilePath.String)
	assert.Equal(t, downloadURLExpiresAt.UnixNano(), report2.DownloadURLExpiresAt.Time.UnixNano())

	report3, err := reportStore.ByPrimaryKey(ctx, report1.UserID, report1.ID)
	require.NoError(t, err)
	assert.Equal(t, report2.ID, report3.ID)
	assert.Equal(t, report2.CreatedAt.UnixNano(), report3.CreatedAt.UnixNano())
	assert.Equal(t, report2.StartedAt.Time.UnixNano(), report3.StartedAt.Time.UnixNano())
	assert.Equal(t, report2.CompletedAt.Time.UnixNano(), report3.CompletedAt.Time.UnixNano())
	assert.Equal(t, report2.FailedAt.Time.UnixNano(), report3.FailedAt.Time.UnixNano())
	assert.Equal(t, errorMsg, report3.ErrorMessage.String)
	assert.Equal(t, downloadURL, report3.DownloadURL.String)
	assert.Equal(t, outputPath, report3.OutputFilePath.String)
	assert.Equal(t, downloadURLExpiresAt.UnixNano(), report3.DownloadURLExpiresAt.Time.UnixNano())
	if err := m.Down(); err != nil {
		require.NoError(t, err)
	}
}
