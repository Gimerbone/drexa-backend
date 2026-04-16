package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"drexa/internal/auth"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// tokenService is the concrete implementation of TokenService.
type tokenService struct {
	secretKey            []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
	issuer               string
}

// jwtClaims is the internal jwt.Claims-compatible struct used for signing/parsing.
type jwtClaims struct {
	UserID          string `json:"user_id"`
	Email           string `json:"email"`
	IsEmailVerified bool   `json:"is_email_verified"`
	IsPhoneVerified bool   `json:"is_phone_verified"`
	IsKycVerified   bool   `json:"is_kyc_verified"`
	jwt.RegisteredClaims
}

// NewTokenService constructs a TokenService.
//
// secretKey    — HMAC-SHA256 signing secret; load from env, never hardcode.
// issuer       — identifies this service in the JWT "iss" claim (e.g. "myapp.api").
// accessTTL    — recommended: 15m for production.
// refreshTTL   — recommended: 7d–30d for production.
func NewTokenService(
	secretKey []byte,
	issuer string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) auth.TokenService {
	return &tokenService{
		secretKey:            secretKey,
		issuer:               issuer,
		accessTokenDuration:  accessTTL,
		refreshTokenDuration: refreshTTL,
	}
}

// ── GenerateAccessToken ──────────────────────────────────────────────────────

func (s *tokenService) GenerateAccessToken(_ context.Context, user *auth.User) (string, error) {
	if user == nil {
		return "", errors.New("token_service: user must not be nil")
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.accessTokenDuration)

	claims := jwtClaims{
		UserID:          user.UserID,
		Email:           user.Email,
		IsEmailVerified: user.IsEmailVerified,
		IsPhoneVerified: user.IsPhoneVerified,
		IsKycVerified:   user.KycProfile.DukcapilVerified, // gate trading endpoints without a DB hit
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   user.UserID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("token_service: signing access token: %w", err)
	}
	return signed, nil
}

// ── GenerateRefreshToken ─────────────────────────────────────────────────────

// GenerateRefreshToken returns a cryptographically random, URL-safe opaque token.
// Store only the hash (via HashToken) in the database — never the raw value.
func (s *tokenService) GenerateRefreshToken(_ context.Context, userID string) (string, error) {
	if userID == "" {
		return "", errors.New("token_service: userID must not be empty")
	}

	// 32 random bytes → 256 bits of entropy, base64url-encoded (~43 chars, no padding)
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("token_service: generating refresh token entropy: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// ── ValidateAccessToken ──────────────────────────────────────────────────────

func (s *tokenService) ValidateAccessToken(_ context.Context, rawToken string) (*auth.JWTClaims, error) {
	parsed, err := jwt.ParseWithClaims(
		rawToken,
		&jwtClaims{},
		func(t *jwt.Token) (interface{}, error) {
			// Reject any token that was signed with an unexpected algorithm.
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("token_service: unexpected signing method: %v", t.Header["alg"])
			}
			return s.secretKey, nil
		},
		jwt.WithIssuer(s.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("token_service: parsing token: %w", err)
	}

	claims, ok := parsed.Claims.(*jwtClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("token_service: invalid token claims")
	}

	return &auth.JWTClaims{
		UserID:          claims.UserID,
		Email:           claims.Email,
		IsEmailVerified: claims.IsEmailVerified,
		IsPhoneVerified: claims.IsPhoneVerified,
		IsKycVerified:   claims.IsKycVerified,
		ExpiresAt:       claims.ExpiresAt.Time,
		IssuedAt:        claims.IssuedAt.Time,
	}, nil
}

// ── HashToken ────────────────────────────────────────────────────────────────

// HashToken produces a hex-encoded SHA-256 digest.
// Use this before persisting any refresh or password-reset token.
// The raw token travels only over the wire; the hash lives in the DB.
func (s *tokenService) HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
