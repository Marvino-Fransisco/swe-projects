package util

import (
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// TokenConfig holds the configuration for JWT token generation.
type TokenConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
}

// DefaultTokenConfig returns the default JWT configuration.
// Override individual fields before passing to NewJWTService.
func DefaultTokenConfig() *TokenConfig {
	return &TokenConfig{
		AccessTokenSecret:  "access-secret-key",
		RefreshTokenSecret: "refresh-secret-key",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
	}
}

// CustomClaims holds the custom JWT claims with user_id and role.
type CustomClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwtv5.RegisteredClaims
}

// JWTService provides JWT token operations.
type JWTService struct {
	config *TokenConfig
}

// NewJWTService creates a new JWT service with the provided configuration.
func NewJWTService(cfg *TokenConfig) *JWTService {
	return &JWTService{
		config: cfg,
	}
}

// GenerateAccessToken creates a new signed access token with the given user_id and role.
func (s *JWTService) GenerateAccessToken(userID, role string) (string, error) {
	now := time.Now()
	claims := CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwtv5.RegisteredClaims{
			IssuedAt:  jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(now.Add(s.config.AccessTokenTTL)),
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.AccessTokenSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken creates a new signed refresh token with the given user_id and role.
func (s *JWTService) GenerateRefreshToken(userID, role string) (string, error) {
	now := time.Now()
	claims := CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwtv5.RegisteredClaims{
			IssuedAt:  jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(now.Add(s.config.RefreshTokenTTL)),
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.RefreshTokenSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// ValidateAccessToken validates an access token and returns the parsed claims.
func (s *JWTService) ValidateAccessToken(tokenString string) (*CustomClaims, error) {
	claims, err := s.parseToken(tokenString, s.config.AccessTokenSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %w", err)
	}
	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns the parsed claims.
func (s *JWTService) ValidateRefreshToken(tokenString string) (*CustomClaims, error) {
	claims, err := s.parseToken(tokenString, s.config.RefreshTokenSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}
	return claims, nil
}

// parseToken parses and validates a token string using the provided secret.
func (s *JWTService) parseToken(tokenString, secret string) (*CustomClaims, error) {
	token, err := jwtv5.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwtv5.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// GetUserID extracts the user_id from a validated access token.
func (s *JWTService) GetUserID(tokenString string) (string, error) {
	claims, err := s.ValidateAccessToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// GetRole extracts the role from a validated access token.
func (s *JWTService) GetRole(tokenString string) (string, error) {
	claims, err := s.ValidateAccessToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.Role, nil
}

// GetClaims extracts all custom claims from a validated access token.
func (s *JWTService) GetClaims(tokenString string) (*CustomClaims, error) {
	return s.ValidateAccessToken(tokenString)
}

// RefreshTokens validates a refresh token and generates a new access/refresh token pair.
func (s *JWTService) RefreshTokens(refreshTokenString string) (string, string, error) {
	claims, err := s.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", fmt.Errorf("refresh token validation failed: %w", err)
	}

	accessToken, err := s.GenerateAccessToken(claims.UserID, claims.Role)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	newRefreshToken, err := s.GenerateRefreshToken(claims.UserID, claims.Role)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}
