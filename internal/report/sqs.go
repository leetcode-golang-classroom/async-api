package report

import "github.com/google/uuid"

type SQSMessage struct {
	UserID   uuid.UUID `json:"user_id"`
	ReportID uuid.UUID `json:"report_id"`
}
