package usecase

import (
	"context"
	"drexa/internal/auth"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type authUsecase struct {
	userRepo         auth.UserRepository
	authProviderRepo auth.AuthProviderRepository
	refreshTokenRepo auth.RefreshTokenRepository
	resetTokenRepo   auth.PasswordResetTokenRepository
	otpService       auth.OTPService
	notifService     auth.NotificationService
	tokenService     auth.TokenService
}

// func NewAuthUsecase(
// 	userRepo auth.UserRepository,
// 	authProviderRepo auth.AuthProviderRepository,
// 	refreshTokenRepo auth.RefreshTokenRepository,
// 	resetTokenRepo auth.PasswordResetTokenRepository,
// 	otpService auth.OTPService,
// 	notifService auth.NotificationService,
// 	tokenService auth.TokenService,
// ) auth.AuthUsecase {
// 	return &authUsecase{
// 		userRepo:         userRepo,
// 		authProviderRepo: authProviderRepo,
// 		refreshTokenRepo: refreshTokenRepo,
// 		resetTokenRepo:   resetTokenRepo,
// 		otpService:       otpService,
// 		notifService:     notifService,
// 		tokenService:     tokenService,
// 	}
// }

// TODO : Implement all usecases

// RequestPasswordReset — uses userRepo + resetTokenRepo + tokenService + notifService
func (uc *authUsecase) RequestPasswordReset(ctx context.Context, email string) error {
	// Always return nil regardless of outcome — never confirm whether the email
	// exists in the system, prevents user enumeration attacks

	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil // silently return even if user not found
	}

	// Clean up stale tokens before issuing a new one — prevents token table bloat
	// and ensures only one valid reset token exists per user at a time
	_ = uc.resetTokenRepo.DeleteExpiredByUserID(ctx, user.UserID)

	// Generate a long random token — tokenService handles crypto/rand generation
	// This is NOT an OTP — it's a full random token used in a reset link
	rawToken, err := uc.tokenService.GenerateRefreshToken(ctx, user.UserID)
	if err != nil {
		return nil // still silent
	}

	// Persist the hash — never store raw tokens
	resetToken := &auth.PasswordResetToken{
		TokenID:   uuid.NewString(),
		UserID:    user.UserID,
		TokenHash: uc.tokenService.HashToken(rawToken),
		ExpiresAt: time.Now().Add(1 * time.Hour), // short window — standard for password resets
	}
	if err := uc.resetTokenRepo.Create(ctx, resetToken); err != nil {
		return nil // still silent
	}

	// Send the raw token to the user's email — notifService builds the full reset URL
	_ = uc.notifService.SendPasswordReset(ctx, user.UserID, user.Email, rawToken)

	return nil
}

// ResetPassword — uses resetTokenRepo + userRepo + refreshTokenRepo + notifService
func (uc *authUsecase) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	// 1. Hash the incoming token to look it up — repo checks used_at and expires_at
	tokenHash := uc.tokenService.HashToken(rawToken)

	stored, err := uc.resetTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		// Covers: not found, already used, or expired — all return the same error
		// so attackers can't distinguish between cases
		return auth.ErrTokenInvalid
	}

	// 2. Hash the new password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 3. Update password
	if err := uc.userRepo.UpdatePasswordHash(ctx, stored.UserID, string(hash)); err != nil {
		return err
	}

	// 4. Consume the reset token immediately — can't be reused even if the link is clicked again
	if err := uc.resetTokenRepo.Revoke(ctx, stored.TokenID); err != nil {
		return err
	}

	// 5. Revoke all active sessions — force re-login on all devices after password change
	// If someone's account was compromised, this kicks out the attacker too
	_ = uc.refreshTokenRepo.RevokeAllByUserID(ctx, stored.UserID)

	// 6. Notify user — security alert so they know their password changed
	user, err := uc.userRepo.FindByID(ctx, stored.UserID)
	if err == nil {
		_ = uc.notifService.SendPasswordChanged(ctx, user.UserID, user.Email)
	}

	return nil
}
