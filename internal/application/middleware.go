package application

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/jwt"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/logger"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/user"
)

func NewLoggerMiddleware(ctx context.Context) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger.FromContext(ctx)
			log.Info("http request", slog.String("path", fmt.Sprintf("%s %s", r.Method, r.URL.Path)))
			next.ServeHTTP(w, r)
		})
	}
}

type userCtxKey struct{}

func ContextWithUserID(ctx context.Context, user *user.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, user)
}

func NewAuthMiddleware(ctx context.Context, jwtManager *jwt.JWTManager, userStore *user.UserStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger.FromContext(ctx)
			if strings.HasPrefix(r.URL.Path, "/auth") {
				next.ServeHTTP(w, r)
				return
			}
			// authorization header
			authHeader := r.Header.Get("Authorization")
			var token string
			if parts := strings.Split(authHeader, "Bearer "); len(parts) == 2 {
				token = parts[1]
			}
			if token == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			parsedToken, err := jwtManager.Parse(token)
			if err != nil {
				log.Error("fialed to parse token", slog.Any("error", err))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if !jwtManager.IsAccessToken(parsedToken) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("not an access token"))
				return
			}

			userIDStr, err := parsedToken.Claims.GetSubject()
			if err != nil {
				log.Error("failed to extract subject claim from token", slog.Any("error", err))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				log.Error("token subject is not valid uuid", slog.Any("error", err))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			user, err := userStore.ByID(r.Context(), userID)
			if err != nil {
				log.Error("failed to get user by id", slog.Any("error", err))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ContextWithUserID(r.Context(), user)))
		})
	}
}
