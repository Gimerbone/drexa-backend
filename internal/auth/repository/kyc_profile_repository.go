package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"drexa/internal/auth"
)

type kycProfileRepository struct {
	db *gorm.DB
}

func NewKycProfileRepository(db *gorm.DB) auth.KycProfileRepository {
	return &kycProfileRepository{db: db}
}

// ─── Write ───────────────────────────────────────────────────────────────────

func (r *kycProfileRepository) Create(ctx context.Context, kyc *auth.KycProfile) error {
	result := r.db.WithContext(ctx).Create(kyc)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			// uniqueIndex on user_id means one KYC profile per user
			return auth.ErrKycAlreadySubmitted
		}
		return result.Error
	}
	return nil
}

func (r *kycProfileRepository) Update(ctx context.Context, kyc *auth.KycProfile) error {
	result := r.db.WithContext(ctx).Save(kyc)
	return result.Error
}

func (r *kycProfileRepository) UpdateStatus(ctx context.Context, kycID string, status auth.KycStatus, reason string) error {
	// Use map[string]any to avoid GORM skipping zero/empty values on struct updates
	result := r.db.WithContext(ctx).
		Model(&auth.KycProfile{}).
		Where("kyc_id = ?", kycID).
		Updates(map[string]any{
			"status":           status,
			"rejection_reason": reason, // empty string is valid here when approving
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return auth.ErrKycNotFound
	}
	return nil
}

// ─── Read ────────────────────────────────────────────────────────────────────

func (r *kycProfileRepository) FindByID(ctx context.Context, kycID string) (*auth.KycProfile, error) {
	var kyc auth.KycProfile
	result := r.db.WithContext(ctx).
		Where("kyc_id = ?", kycID).
		First(&kyc)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, auth.ErrKycNotFound
		}
		return nil, result.Error
	}
	return &kyc, nil
}

func (r *kycProfileRepository) FindByUserID(ctx context.Context, userID string) (*auth.KycProfile, error) {
	var kyc auth.KycProfile
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&kyc)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, auth.ErrKycNotFound
		}
		return nil, result.Error
	}
	return &kyc, nil
}

func (r *kycProfileRepository) FindByStatus(ctx context.Context, status auth.KycStatus) ([]auth.KycProfile, error) {
	var profiles []auth.KycProfile
	result := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("submitted_at ASC"). // oldest first — fair review queue ordering
		Find(&profiles)
	if result.Error != nil {
		return nil, result.Error
	}
	return profiles, nil
}
