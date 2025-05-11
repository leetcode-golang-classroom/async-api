package jwt_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/config"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/jwt"
	"github.com/stretchr/testify/require"
)

func TestJWTManager(t *testing.T) {
	resultConfig := config.AppConfig

	jwtManager := jwt.NewJWTManager(resultConfig)
	userID := uuid.New()
	tokenPair, err := jwtManager.GenerateTokenPair(userID)
	require.NoError(t, err)

	require.True(t, jwtManager.IsAccessToken(tokenPair.AccessToken))
	require.False(t, jwtManager.IsAccessToken(tokenPair.RefreshToken))

	accessTokenIssuer, err := tokenPair.AccessToken.Claims.GetIssuer()
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("http://%s:%s", resultConfig.JWTServerHost, resultConfig.Port), accessTokenIssuer)

	accessTokenSubject, err := tokenPair.AccessToken.Claims.GetSubject()
	require.NoError(t, err)
	require.Equal(t, userID.String(), accessTokenSubject)

	refreshTokenSubject, err := tokenPair.RefreshToken.Claims.GetSubject()
	require.NoError(t, err)
	require.Equal(t, userID.String(), refreshTokenSubject)

	refreshTokenIssuer, err := tokenPair.RefreshToken.Claims.GetIssuer()
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("http://%s:%s", resultConfig.JWTServerHost, resultConfig.Port), refreshTokenIssuer)

	parsedAccessToken, err := jwtManager.Parse(tokenPair.AccessToken.Raw)
	require.NoError(t, err)
	require.Equal(t, tokenPair.AccessToken.Raw, parsedAccessToken.Raw)

	parsedRefreshToken, err := jwtManager.Parse(tokenPair.RefreshToken.Raw)
	require.NoError(t, err)
	require.Equal(t, tokenPair.RefreshToken.Raw, parsedRefreshToken.Raw)
}
