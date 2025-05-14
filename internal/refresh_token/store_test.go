package refreshtoken_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/config"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/db"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/jwt"
	refreshtoken "github.com/leetcode-golang-classroom/golang-async-api/internal/refresh_token"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/user"
	"github.com/stretchr/testify/require"
)

func SetupDB(t *testing.T) (*sql.DB, *migrate.Migrate, *config.Config) {
	t.Helper()
	appConfig := config.AppConfig
	appConfig.SetupEnv(config.Env_Dev)
	dbURL := appConfig.DBURLTEST
	db, err := db.Connect(dbURL)
	require.NoError(t, err)

	result := strings.Replace(appConfig.PROJECT_ROOT, "/internal/refresh_token", "", 1)
	m, err := migrate.New(
		fmt.Sprintf("file://%s/migrations", result),
		dbURL,
	)
	require.NoError(t, err)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}
	return db, m, appConfig
}

func TestUserStore(t *testing.T) {
	db, m, appConfig := SetupDB(t)
	ctx := context.Background()
	defer db.Close()
	refreshTokenStore := refreshtoken.NewRefreshTokenStore(db)
	userStore := user.NewUserStore(db)
	user1, err := userStore.CreateUser(ctx, "test@email.com", "test")
	require.NoError(t, err)

	jwtManager := jwt.NewJWTManager(appConfig)

	tokenPair, err := jwtManager.GenerateTokenPair(user1.ID)
	require.NoError(t, err)

	refreshTokenRecord, err := refreshTokenStore.Create(ctx, user1.ID, tokenPair.RefreshToken)
	require.NoError(t, err)
	require.Equal(t, user1.ID, refreshTokenRecord.UserID)
	expectedExpiration, err := tokenPair.RefreshToken.Claims.GetExpirationTime()
	require.NoError(t, err)
	require.Equal(t, expectedExpiration.Time.Unix(), refreshTokenRecord.ExpiresAt.Unix())

	refreshTokenRecord1, err := refreshTokenStore.ByPrimaryKey(ctx, user1.ID, tokenPair.RefreshToken)
	require.NoError(t, err)
	require.Equal(t, refreshTokenRecord.UserID, refreshTokenRecord1.UserID)
	require.Equal(t, refreshTokenRecord.HashedToken, refreshTokenRecord1.HashedToken)
	require.Equal(t, refreshTokenRecord.CreatedAt, refreshTokenRecord1.CreatedAt)
	require.Equal(t, refreshTokenRecord.ExpiresAt, refreshTokenRecord1.ExpiresAt)

	if err := m.Down(); err != nil {
		require.NoError(t, err)
	}
	defer m.Drop()
}
