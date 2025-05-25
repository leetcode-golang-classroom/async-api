package report

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type ReportStore struct {
	db *sqlx.DB
}

func NewReportStore(db *sql.DB) *ReportStore {
	return &ReportStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

type Report struct {
	UserID               uuid.UUID      `db:"user_id"`
	ID                   uuid.UUID      `db:"id"`
	ReportType           string         `db:"report_type"`
	OutputFilePath       sql.NullString `db:"output_file_path"`
	DownloadURL          sql.NullString `db:"download_url"`
	DownloadURLExpiresAt sql.NullTime   `db:"download_url_expires_at"`
	ErrorMessage         sql.NullString `db:"error_message"`
	CreatedAt            time.Time      `db:"created_at"`
	StartedAt            sql.NullTime   `db:"started_at"`
	CompletedAt          sql.NullTime   `db:"completed_at"`
	FailedAt             sql.NullTime   `db:"failed_at"`
}

func (s *ReportStore) Create(ctx context.Context, userID uuid.UUID, reportType string) (*Report, error) {
	const prepareStmt = `INSERT INTO reports(user_id, report_type) VALUES ($1, $2) RETURNING *`
	var report Report
	if err := s.db.GetContext(ctx, &report, prepareStmt, userID, reportType); err != nil {
		return nil, fmt.Errorf("failed to insert report: %w", err)
	}
	return &report, nil
}

func (s *ReportStore) Update(ctx context.Context, report *Report) (*Report, error) {
	const prepareStmt = `
UPDATE reports
SET output_file_path = $1,
    download_url = $2,
		download_url_expires_at = $3,
		error_message = $4,
		started_at = $5,
	  completed_at = $6,
		failed_at = $7
WHERE user_id = $8 AND id = $9 RETURNING *;	
	`
	var resultReport Report
	if err := s.db.GetContext(
		ctx,
		&resultReport,
		prepareStmt,
		report.OutputFilePath,
		report.DownloadURL,
		report.DownloadURLExpiresAt,
		report.ErrorMessage,
		report.StartedAt,
		report.CompletedAt,
		report.FailedAt,
		report.UserID,
		report.ID,
	); err != nil {
		return nil, fmt.Errorf("failed to update report %s for user %s: %w", report.ID, report.UserID, err)
	}
	return &resultReport, nil
}

func (s *ReportStore) ByPrimaryKey(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*Report, error) {
	const prepareStmt = `SELECT * FROM reports WHERE user_id = $1 AND id = $2;`
	var report Report
	if err := s.db.GetContext(ctx, &report, prepareStmt, userID, id); err != nil {
		return nil, fmt.Errorf("failed to query report %s for user %s: %w", id, userID, err)
	}
	return &report, nil
}
