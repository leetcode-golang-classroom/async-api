package user_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/config"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/db"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func SetupDB(t *testing.T) (*sql.DB, *migrate.Migrate) {
	t.Helper()
	appConfig := config.AppConfig
	appConfig.SetupEnv(config.Env_Dev)
	dbURL := appConfig.DBURLTEST
	db, err := db.Connect(dbURL)
	require.NoError(t, err)

	result := strings.Replace(appConfig.PROJECT_ROOT, "/internal/user", "", 1)
	m, err := migrate.New(
		fmt.Sprintf("file://%s/migrations", result),
		dbURL,
	)
	require.NoError(t, err)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}
	return db, m
}

func TestUserStore(t *testing.T) {
	db, m := SetupDB(t)
	defer db.Close()
	userStore := user.NewUserStore(db)
	now := time.Now()
	ctx := context.Background()
	user1, err := userStore.CreateUser(ctx, "test@test.com", "testpassword")
	require.NoError(t, err)
	assert.Equal(t, "test@test.com", user1.Email)
	require.NoError(t, user1.ComparePassword("testpassword"))
	assert.Less(t, now.UnixNano(), user1.CreatedAt.UnixNano())

	user2, err := userStore.ByID(ctx, user1.ID)
	require.NoError(t, err)
	assert.Equal(t, user1.Email, user2.Email)
	assert.Equal(t, user1.ID, user2.ID)
	assert.Equal(t, user1.HashedPasswordBase64, user2.HashedPasswordBase64)
	assert.Equal(t, user1.CreatedAt.UnixNano(), user2.CreatedAt.UnixNano())

	user2, err = userStore.ByEmail(ctx, user1.Email)
	require.NoError(t, err)
	assert.Equal(t, user1.Email, user2.Email)
	assert.Equal(t, user1.ID, user2.ID)
	assert.Equal(t, user1.HashedPasswordBase64, user2.HashedPasswordBase64)
	assert.Equal(t, user1.CreatedAt.UnixNano(), user2.CreatedAt.UnixNano())
	if err := m.Down(); err != nil {
		require.NoError(t, err)
	}
}
