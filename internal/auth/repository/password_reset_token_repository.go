package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"drexa/internal/auth"
)

type passwordResetTokenRepository struct {
	db *gorm.DB
}

func NewPasswordResetTokenRepository(db *gorm.DB) auth.PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

// ─── Write ───────────────────────────────────────────────────────────────────

func (r *passwordResetTokenRepository) Create(ctx context.Context, token *auth.PasswordResetToken) error {
	result := r.db.WithContext(ctx).Create(token)
	return result.Error
}

func (r *passwordResetTokenRepository) Revoke(ctx context.Context, tokenID string) error {
	// Mark as used rather than deleting — keeps an audit record that the token was redeemed
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&auth.PasswordResetToken{}).
		Where("token_id = ? AND used_at IS NULL", tokenID).
		Update("used_at", &now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// Either token doesn't exist or was already used — both map to invalid
		return auth.ErrTokenInvalid
	}
	return nil
}

func (r *passwordResetTokenRepository) DeleteExpiredByUserID(ctx context.Context, userID string) error {
	// Called at the start of RequestPasswordReset to clean up stale tokens
	// before issuing a new one — ensures only one valid token per user at a time
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND expires_at < ?", userID, time.Now()).
		Delete(&auth.PasswordResetToken{})
	return result.Error
}

// ─── Read ────────────────────────────────────────────────────────────────────

func (r *passwordResetTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*auth.PasswordResetToken, error) {
	var token auth.PasswordResetToken
	result := r.db.WithContext(ctx).
		Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, time.Now()).
		First(&token)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Don't distinguish between not found, already used, or expired —
			// all map to the same error so attackers can't probe token states
			return nil, auth.ErrTokenInvalid
		}
		return nil, result.Error
	}
	return &token, nil
}
