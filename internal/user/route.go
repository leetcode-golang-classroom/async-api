package user

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/helper"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/jwt"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/response"
	refreshtoken "github.com/leetcode-golang-classroom/golang-async-api/internal/refresh_token"
)

type Handler struct {
	logger            *slog.Logger
	validator         *validator.Validate
	userStore         *UserStore
	refreshTokenStore *refreshtoken.RefreshTokenStore
	jwtManager        *jwt.JWTManager
}

func NewHandler(logger *slog.Logger, validator *validator.Validate, userStore *UserStore,
	refreshTokenStore *refreshtoken.RefreshTokenStore,
	jwtManager *jwt.JWTManager,
) *Handler {
	return &Handler{
		logger:            logger,
		validator:         validator,
		userStore:         userStore,
		refreshTokenStore: refreshTokenStore,
		jwtManager:        jwtManager,
	}
}

func (h *Handler) RegisterRoute(router *http.ServeMux) {
	// setup route
	router.HandleFunc("POST /auth/signup", h.signUpHandler())
	router.HandleFunc("POST /auth/signIn", h.signInHandler())
}

func (h *Handler) signUpHandler() http.HandlerFunc {
	return helper.Handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := helper.Decode[SignUpRequest](r, h.validator)
		if err != nil {
			return helper.NewErrWithStatus(
				http.StatusBadRequest,
				err,
			)
		}
		defer r.Body.Close()
		// find existed user
		existingUser, err := h.userStore.ByEmail(r.Context(), req.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return helper.NewErrWithStatus(
				http.StatusInternalServerError,
				err,
			)
		}
		// request user existed
		if existingUser != nil {
			return helper.NewErrWithStatus(
				http.StatusConflict,
				fmt.Errorf("email already exists: %v", existingUser.Email),
			)
		}
		// create user
		if _, err := h.userStore.CreateUser(r.Context(), req.Email, req.Password); err != nil {
			return helper.NewErrWithStatus(
				http.StatusInternalServerError,
				err,
			)
		}
		// response with successfully signed up
		if err := helper.Encode(response.ApiResponse[struct{}]{
			Message: "successfully signed up user",
		}, http.StatusCreated, w); err != nil {
			return helper.NewErrWithStatus(
				http.StatusInternalServerError,
				err,
			)
		}
		return nil
	})
}

func (h *Handler) signInHandler() http.HandlerFunc {
	return helper.Handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := helper.Decode[SignInRequest](r, h.validator)
		if err != nil {
			return helper.NewErrWithStatus(
				http.StatusBadRequest,
				err,
			)
		}
		defer r.Body.Close()
		user, err := h.userStore.ByEmail(r.Context(), req.Email)
		if err != nil {
			return helper.NewErrWithStatus(
				http.StatusInternalServerError,
				err,
			)
		}
		if err := user.ComparePassword(req.Password); err != nil {
			return helper.NewErrWithStatus(
				http.StatusUnauthorized,
				err,
			)
		}
		tokenPair, err := h.jwtManager.GenerateTokenPair(user.ID)
		if err != nil {
			return helper.NewErrWithStatus(
				http.StatusInternalServerError,
				err,
			)
		}
		// store refreshToken
		_, err = h.refreshTokenStore.ResetUserToken(r.Context(), user.ID, tokenPair.RefreshToken)
		if err != nil {
			return helper.NewErrWithStatus(
				http.StatusInternalServerError,
				err,
			)
		}

		if err := helper.Encode(response.ApiResponse[SignInResponse]{
			Data: &SignInResponse{
				AccessToken:  tokenPair.AccessToken.Raw,
				RefreshToken: tokenPair.RefreshToken.Raw,
			},
		}, http.StatusOK, w); err != nil {
			return helper.NewErrWithStatus(
				http.StatusInternalServerError,
				err,
			)
		}
		return nil
	})
}
