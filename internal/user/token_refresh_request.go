package user

import "github.com/go-playground/validator/v10"

type TokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (r TokenRefreshRequest) Validate(validator *validator.Validate) error {
	err := validator.Struct(r)
	if err != nil {
		return err
	}
	return nil
}

type TokenRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
