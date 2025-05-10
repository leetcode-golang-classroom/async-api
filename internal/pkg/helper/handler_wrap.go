package helper

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/logger"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/response"
)

type ErrWithStatus struct {
	status int
	err    error
}

func NewErrWithStatus(status int, err error) *ErrWithStatus {
	return &ErrWithStatus{status: status, err: err}
}

func (e *ErrWithStatus) Error() string {
	return e.err.Error()
}

// Handler - handler that will handle error message
func Handler(fn func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// handle error message
		if err := fn(w, r); err != nil {
			status := http.StatusInternalServerError
			msg := http.StatusText(status)
			if e, ok := err.(*ErrWithStatus); ok {
				status = e.status
				msg = http.StatusText(e.status)
				if status == http.StatusBadRequest || status == http.StatusConflict {
					msg = e.err.Error()
				}
			}
			log := logger.FromContext(r.Context())
			log.ErrorContext(r.Context(),
				"error executing handler",
				slog.Any("err", err),
				slog.Int("status", status),
				slog.String("msg", msg),
			)
			w.Header().Set("Content-Type", "application/json;charset=utf-8")
			w.WriteHeader(status)
			if err := json.NewEncoder(w).Encode(response.ApiResponse[struct{}]{
				Message: msg,
			}); err != nil {
				log.ErrorContext(r.Context(), "error encoding response", slog.Any("err", err))
			}
		}
	}
}
