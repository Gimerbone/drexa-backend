package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"drexa/internal/auth"
)

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) auth.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

// ─── Write ───────────────────────────────────────────────────────────────────

func (r *refreshTokenRepository) Create(ctx context.Context, token *auth.RefreshToken) error {
	result := r.db.WithContext(ctx).Create(token)
	return result.Error
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, tokenID string) error {
	// Set revoked_at instead of deleting — keeps session history for audit purposes
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&auth.RefreshToken{}).
		Where("token_id = ? AND revoked_at IS NULL", tokenID).
		Update("revoked_at", &now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// Either token doesn't exist or was already revoked — both map to invalid
		return auth.ErrTokenInvalid
	}
	return nil
}

func (r *refreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	// Used for "logout from all devices" — sets revoked_at on every active session
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&auth.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", &now)
	return result.Error
}

// ─── Read ────────────────────────────────────────────────────────────────────

func (r *refreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*auth.RefreshToken, error) {
	var token auth.RefreshToken
	result := r.db.WithContext(ctx).
		Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", tokenHash, time.Now()).
		First(&token)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Don't distinguish between not found, revoked, or expired — all map to invalid
			return nil, auth.ErrTokenInvalid
		}
		return nil, result.Error
	}
	return &token, nil
}

func (r *refreshTokenRepository) FindActiveByUserID(ctx context.Context, userID string) ([]auth.RefreshToken, error) {
	// Used for "active sessions" screen — shows user which devices are logged in
	var tokens []auth.RefreshToken
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		Find(&tokens)
	if result.Error != nil {
		return nil, result.Error
	}
	return tokens, nil
}

// ─── Cleanup ─────────────────────────────────────────────────────────────────

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	// Intended to be called by a background cron job — not inline in requests
	// Permanently deletes expired rows to keep the table lean
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&auth.RefreshToken{})
	return result.Error
}
