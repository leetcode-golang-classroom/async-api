package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/config"
)

var signingMethod = jwt.SigningMethodHS256

type JWTManager struct {
	config *config.Config
}

type TokenPair struct {
	AccessToken  *jwt.Token
	RefreshToken *jwt.Token
}

type CustomClaims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func NewJWTManager(config *config.Config) *JWTManager {
	return &JWTManager{
		config: config,
	}
}

// Parse - parse token to token.Claim, so that the custom claim could load setup
func (jwtManager *JWTManager) Parse(token string) (*jwt.Token, error) {
	parser := jwt.NewParser()
	jwtToken, err := parser.Parse(token, func(t *jwt.Token) (any, error) {
		if t.Method != signingMethod {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtManager.config.JWTSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	return jwtToken, nil
}

// GenerateTokenPair - generate accessToken, refreshToken
func (jwtManager *JWTManager) GenerateTokenPair(userID uuid.UUID) (*TokenPair, error) {
	now := time.Now().UTC()
	issuer := fmt.Sprintf("http://%s:%s", jwtManager.config.JWTServerHost, jwtManager.config.Port)
	jwtAccessToken := jwt.NewWithClaims(signingMethod,
		CustomClaims{
			TokenType: "access",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   userID.String(),
				Issuer:    issuer,
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute * 15)),
				IssuedAt:  jwt.NewNumericDate(now),
			},
		})

	key := []byte(jwtManager.config.JWTSecret)
	signedAccessToken, err := jwtAccessToken.SignedString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}
	accessToken, err := jwtManager.Parse(signedAccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse access token: %w", err)
	}

	jwtRefreshToken := jwt.NewWithClaims(signingMethod,
		CustomClaims{
			TokenType: "refresh",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   userID.String(),
				Issuer:    issuer,
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * 24 * 30)),
				IssuedAt:  jwt.NewNumericDate(now),
			},
		})
	signedRefreshToken, err := jwtRefreshToken.SignedString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}
	refreshToken, err := jwtManager.Parse(signedRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil

}

// IsAccessToken - check if this token is access token
func (jwtManager *JWTManager) IsAccessToken(token *jwt.Token) bool {
	jwtClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	if tokenType, ok := jwtClaims["token_type"]; ok {
		return tokenType == "access"
	}
	return false
}
