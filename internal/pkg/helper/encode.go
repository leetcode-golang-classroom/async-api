package helper

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Encode - encode response body
func Encode[T any](v T, status int, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encoding response: %w", err)
	}
	return nil
}
