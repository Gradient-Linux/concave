package auth

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	DefaultTokenTTL   = 24 * time.Hour
	DefaultRefreshCut = 25
	SessionCookieName = "concave_session"
)

// Claims holds the JWT payload.
type Claims struct {
	Role Role `json:"role"`
	jwt.RegisteredClaims
}

// TokenConfig holds the signing key and TTL.
type TokenConfig struct {
	SigningKey []byte        `json:"signing_key"`
	TokenTTL   time.Duration `json:"token_ttl"`
}

// LoadOrCreateTokenConfig loads or creates the auth signing config.
func LoadOrCreateTokenConfig(workspaceRoot string) (TokenConfig, error) {
	path := tokenConfigPath(workspaceRoot)
	data, err := os.ReadFile(path)
	if err == nil {
		var cfg TokenConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return TokenConfig{}, fmt.Errorf("decode %s: %w", path, err)
		}
		if cfg.TokenTTL <= 0 {
			cfg.TokenTTL = DefaultTokenTTL
		}
		return cfg, nil
	}
	if !os.IsNotExist(err) {
		return TokenConfig{}, fmt.Errorf("read %s: %w", path, err)
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return TokenConfig{}, fmt.Errorf("generate signing key: %w", err)
	}
	cfg := TokenConfig{
		SigningKey: key,
		TokenTTL:   DefaultTokenTTL,
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return TokenConfig{}, fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return TokenConfig{}, fmt.Errorf("encode %s: %w", path, err)
	}
	if err := os.WriteFile(path, append(payload, '\n'), 0o600); err != nil {
		return TokenConfig{}, fmt.Errorf("write %s: %w", path, err)
	}
	return cfg, nil
}

func tokenConfigPath(workspaceRoot string) string {
	if override := os.Getenv("CONCAVE_AUTH_CONFIG_PATH"); override != "" {
		return override
	}
	return filepath.Join(workspaceRoot, "config", "auth.json")
}

// IssueToken issues a signed JWT for the given username and role.
func IssueToken(cfg TokenConfig, username string, role Role) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.ttl())),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(cfg.SigningKey)
}

// ValidateToken parses and validates a JWT.
func ValidateToken(cfg TokenConfig, tokenStr string) (Claims, error) {
	var claims Claims
	parsed, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		return cfg.SigningKey, nil
	}, jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return Claims{}, err
	}
	if !parsed.Valid {
		return Claims{}, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// RefreshToken issues a new token if the current token is near expiry.
func RefreshToken(cfg TokenConfig, tokenStr string) (string, error) {
	claims, err := ValidateToken(cfg, tokenStr)
	if err != nil {
		return "", err
	}
	if claims.ExpiresAt == nil {
		return "", fmt.Errorf("token missing expiry")
	}
	remaining := time.Until(claims.ExpiresAt.Time)
	window := cfg.ttl() / 4
	if remaining > window {
		return "", fmt.Errorf("token not yet eligible for refresh")
	}
	return IssueToken(cfg, claims.Subject, claims.Role)
}

func (cfg TokenConfig) ttl() time.Duration {
	if cfg.TokenTTL <= 0 {
		return DefaultTokenTTL
	}
	return cfg.TokenTTL
}
