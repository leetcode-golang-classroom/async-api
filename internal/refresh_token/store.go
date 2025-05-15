package refreshtoken

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type RefreshTokenStore struct {
	db *sqlx.DB
}

func NewRefreshTokenStore(db *sql.DB) *RefreshTokenStore {
	return &RefreshTokenStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

type RefreshToken struct {
	UserID      uuid.UUID `db:"user_id"`
	HashedToken string    `db:"hashed_token"`
	CreatedAt   time.Time `db:"created_at"`
	ExpiresAt   time.Time `db:"expired_at"`
}

func (*RefreshTokenStore) getBase64HashFromToken(token *jwt.Token) (string, error) {
	h := sha256.New()
	h.Write([]byte(token.Raw))
	hashedBytes := h.Sum(nil)
	base64TokenHash := base64.StdEncoding.EncodeToString(hashedBytes)
	return base64TokenHash, nil
}

func (s *RefreshTokenStore) Create(ctx context.Context, userID uuid.UUID, token *jwt.Token) (*RefreshToken, error) {
	const prepareStmt = `INSERT INTO refresh_tokens(user_id, hashed_token, expired_at) VALUES($1, $2, $3) RETURNING *;`
	base64TokenHash, err := s.getBase64HashFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to get base64 encoded token hash: %w", err)
	}
	var refreshToken RefreshToken
	expiresAt, err := token.Claims.GetExpirationTime()
	if err != nil {
		return nil, fmt.Errorf("failed to get expiration time from token: %w", err)
	}
	if err := s.db.GetContext(ctx, &refreshToken, prepareStmt, userID, base64TokenHash, expiresAt.Time.UTC()); err != nil {
		return nil, fmt.Errorf("failed to create refresh token record: %w", err)
	}

	return &refreshToken, nil
}

func (s *RefreshTokenStore) ByPrimaryKey(ctx context.Context, userID uuid.UUID, token *jwt.Token) (*RefreshToken, error) {
	const prepareStmt = `SELECT * FROM refresh_tokens WHERE user_id = $1 AND hashed_token = $2;`
	var refreshToken RefreshToken
	base64TokenHash, err := s.getBase64HashFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to get base64 encoded token hash: %w", err)
	}
	if err := s.db.GetContext(ctx, &refreshToken, prepareStmt,
		userID, base64TokenHash,
	); err != nil {
		return nil, fmt.Errorf("failed to read refresh token record: %w", err)
	}
	return &refreshToken, nil
}

func (s *RefreshTokenStore) DeleteUserToken(ctx context.Context, userID uuid.UUID) (sql.Result, error) {
	const prepareStmt = `DELETE FROM refresh_tokens WHERE user_id = $1;`
	result, err := s.db.ExecContext(ctx, prepareStmt, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete refresh_tokens record: %w", err)
	}
	return result, nil
}
