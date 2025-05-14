package user

import (
	"github.com/go-playground/validator/v10"
)

type SignUpRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (r SignUpRequest) Validate(validator *validator.Validate) error {
	err := validator.Struct(r)
	if err != nil {
		return err
	}
	return nil
}

type SignInRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (r SignInRequest) Validate(validator *validator.Validate) error {
	err := validator.Struct(r)
	if err != nil {
		return err
	}
	return nil
}

type SignInResponse struct {
	AccessToken  string `access_token`
	RefreshToken string `refresh_token`
}
