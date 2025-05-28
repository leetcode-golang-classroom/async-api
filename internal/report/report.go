package report

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateReportRequest struct {
	ReportType string `json:"report_type" validate:"required"`
}

func (r CreateReportRequest) Validate(validator *validator.Validate) error {
	err := validator.Struct(r)
	if err != nil {
		return err
	}
	return nil
}

type ApiReport struct {
	ID                   uuid.UUID  `json:"id"`
	ReportType           string     `json:"report_type"`
	OutputFilePath       *string    `json:"output_file_path,omitempty"`
	DownloadURL          *string    `json:"download_url,omitempty"`
	DownloadURLExpiresAt *time.Time `json:"download_url_expires_at,omitempty"`
	ErrorMessage         *string    `json:"error_message,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	StartedAt            *time.Time `json:"started_at,omitempty"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	FailedAt             *time.Time `json:"failed_at,omitempty"`
	Status               string     `json:"status,omitempty"`
}
