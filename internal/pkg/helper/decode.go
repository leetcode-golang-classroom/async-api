package helper

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type Validator interface {
	Validate(_validator *validator.Validate) error
}

// Decode - decode and vaildate input request body
func Decode[T Validator](r *http.Request, _validator *validator.Validate) (T, error) {
	var t T
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return t, fmt.Errorf("decoding request body: %w", err)
	}
	if err := t.Validate(_validator); err != nil {
		return t, err
	}
	return t, nil
}
