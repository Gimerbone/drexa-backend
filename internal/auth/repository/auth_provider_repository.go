package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"drexa/internal/auth"
)

type authProviderRepository struct {
	db *gorm.DB
}

func NewAuthProviderRepository(db *gorm.DB) auth.AuthProviderRepository {
	return &authProviderRepository{db: db}
}

// ─── Write ───────────────────────────────────────────────────────────────────

func (r *authProviderRepository) Create(ctx context.Context, authProvider *auth.AuthProvider) error {
	result := r.db.WithContext(ctx).Create(authProvider)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return auth.ErrAuthProviderAlreadyExists
		}
		return result.Error
	}
	return nil
}

func (r *authProviderRepository) Delete(ctx context.Context, authID string) error {
	result := r.db.WithContext(ctx).
		Where("auth_id = ?", authID).
		Delete(&auth.AuthProvider{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return auth.ErrAuthProviderNotFound
	}
	return nil
}

// ─── Read ────────────────────────────────────────────────────────────────────

func (r *authProviderRepository) FindByID(ctx context.Context, authID string) (*auth.AuthProvider, error) {
	var authProvider auth.AuthProvider
	result := r.db.WithContext(ctx).
		Where("auth_id = ?", authID).
		First(&authProvider)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, auth.ErrAuthProviderNotFound
		}
		return nil, result.Error
	}
	return &authProvider, nil
}

func (r *authProviderRepository) FindByUserID(ctx context.Context, userID string) ([]auth.AuthProvider, error) {
	var authProviders []auth.AuthProvider
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&authProviders)
	if result.Error != nil {
		return nil, result.Error
	}
	// Find() never returns ErrRecordNotFound — an empty slice is valid (user has no linked providers yet)
	return authProviders, nil
}

func (r *authProviderRepository) FindByProvider(ctx context.Context, provider, providerUID string) (*auth.AuthProvider, error) {
	var authProvider auth.AuthProvider
	result := r.db.WithContext(ctx).
		Where("provider = ? AND provider_uid = ?", provider, providerUID).
		First(&authProvider)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, auth.ErrAuthProviderNotFound
		}
		return nil, result.Error
	}
	return &authProvider, nil
}

func (r *authProviderRepository) ExistsByProvider(ctx context.Context, provider, providerUID string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&auth.AuthProvider{}).
		Where("provider = ? AND provider_uid = ?", provider, providerUID).
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}
