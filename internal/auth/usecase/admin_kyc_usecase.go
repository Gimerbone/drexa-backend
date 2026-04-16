package usecase

import (
	"context"
	"drexa/internal/auth"
)

type adminKyc struct {
	userRepo         auth.UserRepository
	authProviderRepo auth.AuthProviderRepository
	refreshTokenRepo auth.RefreshTokenRepository
	resetTokenRepo   auth.PasswordResetTokenRepository
	otpService       auth.OTPService
	notifService     auth.NotificationService
	tokenService     auth.TokenService
}

func ListByStatus(ctx context.Context, status auth.KycStatus) ([]auth.KycProfile, error) // admin review queue
func GetByID(ctx context.Context, kycID string) (*auth.KycProfile, error)
func GetDecryptedNIK(ctx context.Context, kycID string) (string, error) // decrypts NIK for admin review
func Approve(ctx context.Context, kycID, reviewedBy string) error
func Reject(ctx context.Context, kycID, reviewedBy, reason string) error
func UpdateStatus(ctx context.Context, kycID string, status auth.KycStatus) error
