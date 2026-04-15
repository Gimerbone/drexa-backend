package auth

import "context"

// AuthUsecase handles user-facing authentication flows
type AuthUsecase interface {
	// Registration
	Register(ctx context.Context, email, password string) (*User, error)
	RegisterWithOAuth(ctx context.Context, provider, providerUID, email string) (*User, error)

	// Verification — Send triggers OTP dispatch, Verify checks the code
	SendEmailVerificationOTP(ctx context.Context, userID string) error
	SendPhoneVerificationOTP(ctx context.Context, userID string) error
	VerifyEmail(ctx context.Context, userID, otp string) (bool, error)
	VerifyPhone(ctx context.Context, userID, otp string) (bool, error)

	// Auth
	Login(ctx context.Context, email, password string) (*AuthToken, error)
	LoginWithOAuth(ctx context.Context, provider, providerUID string) (*AuthToken, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthToken, error)
	Logout(ctx context.Context, tokenID string) error
	LogoutAll(ctx context.Context, userID string) error // revokes all sessions across devices

	// Password
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
	RequestPasswordReset(ctx context.Context, email string) error // always returns nil — never confirm email existence
	ResetPassword(ctx context.Context, token, newPassword string) error

	// Trading PIN — separate from login auth, required before executing trades or withdrawals
	SetTradingPin(ctx context.Context, userID, pin string) error
	VerifyTradingPin(ctx context.Context, userID, pin string) (bool, error)
}

// AuthProviderUsecase handles OAuth provider linking and unlinking
type AuthProviderUsecase interface {
	LinkAuthProvider(ctx context.Context, userID, provider, providerUID string) (*AuthProvider, error)
	UnlinkAuthProvider(ctx context.Context, userID, authID string) error // should block if it's the only auth method
	GetAuthMethods(ctx context.Context, userID string) ([]AuthProvider, error)
	FindByProvider(ctx context.Context, provider, providerUID string) (*AuthProvider, error)
}

// KycUsecase handles user-facing KYC submission and status checks
type KycUsecase interface {
	Submit(ctx context.Context, userID string, kyc *KycProfile) error
	GetByUserID(ctx context.Context, userID string) (*KycProfile, error) // user checks their own status
	IsVerified(ctx context.Context, userID string) (bool, error)
	IsExpired(ctx context.Context, userID string) (bool, error)
}

// AdminKycUsecase handles admin-facing KYC review operations
type AdminKycUsecase interface {
	ListByStatus(ctx context.Context, status KycStatus) ([]KycProfile, error) // admin review queue
	GetByID(ctx context.Context, kycID string) (*KycProfile, error)
	GetDecryptedNIK(ctx context.Context, kycID string) (string, error) // decrypts NIK for admin review
	Approve(ctx context.Context, kycID, reviewedBy string) error
	Reject(ctx context.Context, kycID, reviewedBy, reason string) error
	UpdateStatus(ctx context.Context, kycID string, status KycStatus) error
}
