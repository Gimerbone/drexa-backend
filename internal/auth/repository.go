package auth

import "context"

// UserRepository handles persistence for User entities
type UserRepository interface {
	// Write
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, userID string) error // soft deletes via gorm.DeletedAt

	// Read
	FindByID(ctx context.Context, userID string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// Targeted updates — use column-specific updates to avoid GORM zero-value pitfalls
	UpdateEmailVerified(ctx context.Context, userID string, verified bool) error
	UpdatePhoneVerified(ctx context.Context, userID string, verified bool) error
	UpdateLastLoginAt(ctx context.Context, userID string) error
	UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error
	UpdateTradingPinHash(ctx context.Context, userID, pinHash string) error
}

// AuthProviderRepository handles persistence for OAuth provider links
type AuthProviderRepository interface {
	// Write
	Create(ctx context.Context, authProvider *AuthProvider) error
	Delete(ctx context.Context, authID string) error

	// Read
	FindByID(ctx context.Context, authID string) (*AuthProvider, error)
	FindByUserID(ctx context.Context, userID string) ([]AuthProvider, error)
	FindByProvider(ctx context.Context, provider, providerUID string) (*AuthProvider, error)
	ExistsByProvider(ctx context.Context, provider, providerUID string) (bool, error)
}

// RefreshTokenRepository handles persistence for refresh token sessions
type RefreshTokenRepository interface {
	// Write
	Create(ctx context.Context, token *RefreshToken) error
	Revoke(ctx context.Context, tokenID string) error           // single session logout
	RevokeAllByUserID(ctx context.Context, userID string) error // logout from all devices

	// Read
	FindByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	FindActiveByUserID(ctx context.Context, userID string) ([]RefreshToken, error) // for "active sessions" screen

	// Cleanup — intended to be called by a background cron job, not inline in requests
	DeleteExpired(ctx context.Context) error
}

// PasswordResetTokenRepository handles persistence for password reset tokens
type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *PasswordResetToken) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	Revoke(ctx context.Context, tokenID string) error
	DeleteExpiredByUserID(ctx context.Context, userID string) error // cleanup stale tokens on new request
}

// KycProfileRepository handles persistence for KYC profiles
type KycProfileRepository interface {
	// Write
	Create(ctx context.Context, kyc *KycProfile) error
	Update(ctx context.Context, kyc *KycProfile) error
	UpdateStatus(ctx context.Context, kycID string, status KycStatus, reason string) error

	// Read
	FindByID(ctx context.Context, kycID string) (*KycProfile, error)
	FindByUserID(ctx context.Context, userID string) (*KycProfile, error)
	FindByStatus(ctx context.Context, status KycStatus) ([]KycProfile, error) // used by admin review queue
}
