package usecase

import (
	"context"
	"drexa/internal/auth"
	"errors"
	"fmt"
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

func NewAuthUsecase(
	userRepo auth.UserRepository,
	authProviderRepo auth.AuthProviderRepository,
	refreshTokenRepo auth.RefreshTokenRepository,
	resetTokenRepo auth.PasswordResetTokenRepository,
	otpService auth.OTPService,
	notifService auth.NotificationService,
	tokenService auth.TokenService,
) auth.AuthUsecase {
	return &authUsecase{
		userRepo:         userRepo,
		authProviderRepo: authProviderRepo,
		refreshTokenRepo: refreshTokenRepo,
		resetTokenRepo:   resetTokenRepo,
		otpService:       otpService,
		notifService:     notifService,
		tokenService:     tokenService,
	}
}

// TODO : Implement all usecases
// register
func (uc authUsecase) Register(ctx context.Context, email, password string) (*auth.User, error) {
	var user *auth.User

	isExistEmail, err := uc.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return user, errors.New("error to get email")
	}

	if isExistEmail {
		return user, auth.ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return user, errors.New("hash failed")
	}

	user = &auth.User{
		UserID:          uuid.NewString(),
		Email:           email,
		PasswordHash:    string(hash),
		IsEmailVerified: false,
		IsPhoneVerified: false,
	}

	// if uc.userRepo.Create(ctx, user) != nil {
	// 	return nil, errors.New("failed to create user")
	// }

	// if uc.SendEmailVerificationOTP(ctx, user.UserID) != nil {
	// 	return nil, errors.New("failed to send otp")
	// }

	return user, nil
}

// register with oauth
func (uc *authUsecase) RegisterWithOAuth(ctx context.Context, provider, providerUID, email string) (*auth.User, error) {
	var prov *auth.AuthProvider
	userid := uuid.NewString()
	user := &auth.User{
		UserID:     userid,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	prov = &auth.AuthProvider{
		AuthID:      uuid.NewString(),
		UserID:      userid,
		Provider:    provider,
		ProviderUID: providerUID,
		Email:       email,
		CreatedAt:   time.Now(),
	}

	user.AuthMethods = append(user.AuthMethods, *prov)

	return user, nil
}

// Verification
// perlu perbaikan
func (uc authUsecase) SendEmailVerificationOTP(ctx context.Context, userID string) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if uc.otpService.GenerateAndSendEmail(ctx, fmt.Sprintf("otp:email:%s", user.Email), user.Email) != nil {
		return err
	}
	return nil
}

// perlu perbaikan
func (uc authUsecase) SendPhoneVerificationOTP(ctx context.Context, userID string) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if uc.otpService.GenerateAndSendSMS(ctx, fmt.Sprintf("otp:phone:%s", user.PhoneNumber), user.PhoneNumber) != nil {
		return err
	}
	return nil
}
func (uc authUsecase) VerifyEmail(ctx context.Context, userID, otp string) (bool, error) {
	ok, err := uc.otpService.Verify(ctx, userID, otp)
	if err != nil {
		return false, err
	}
	return ok, nil
}
func (uc authUsecase) VerifyPhone(ctx context.Context, userID, otp string) (bool, error) {
	ok, err := uc.otpService.Verify(ctx, userID, otp)
	if err != nil {
		return false, err
	}
	return ok, nil
}

// Auth
func (uc authUsecase) Login(ctx context.Context, email, password string) (*auth.AuthToken, error) {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, auth.ErrUserNotFound
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, err
	}

	Acctoken, err := uc.tokenService.GenerateAccessToken(ctx, user)
	if err != nil {
		return nil, err
	}

	refToken, err := uc.tokenService.GenerateRefreshToken(ctx, user.UserID)
	if err != nil {
		return nil, err
	}

	token := &auth.AuthToken{
		AccessToken:  Acctoken,
		RefreshToken: refToken,
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(time.Minute * 15),
	}
	return token, nil
}
func (uc authUsecase) LoginWithOAuth(ctx context.Context, provider, providerUID string) (*auth.AuthToken, error) {
	prov, err := uc.authProviderRepo.FindByProvider(ctx, provider, providerUID)
	if err != nil {
		return nil, auth.ErrAuthProviderNotFound
	}

	user, err := uc.userRepo.FindByID(ctx, prov.UserID)
	if err != nil {
		return nil, auth.ErrUserNotFound
	}

	Acctoken, err := uc.tokenService.GenerateAccessToken(ctx, user)
	if err != nil {
		return nil, err
	}

	refToken, err := uc.tokenService.GenerateRefreshToken(ctx, user.UserID)
	if err != nil {
		return nil, err
	}

	token := &auth.AuthToken{
		AccessToken:  Acctoken,
		RefreshToken: refToken,
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(time.Minute * 15),
	}
	return token, nil
}

func (uc authUsecase) RefreshToken(ctx context.Context, refreshToken string) (*auth.AuthToken, error) {
	hash := uc.tokenService.HashToken(refreshToken)
	refToken, err := uc.refreshTokenRepo.FindByTokenHash(ctx, hash)
	if err != nil {
		return nil, auth.ErrTokenInvalid
	}

	user, err := uc.userRepo.FindByID(ctx, refToken.UserID)
	if err != nil {
		return nil, auth.ErrUserNotFound
	}

	accToken, err := uc.tokenService.GenerateAccessToken(ctx, user)
	if err != nil {
		return nil, err
	}

	NewRefToken, err := uc.tokenService.GenerateRefreshToken(ctx, user.UserID)
	if err != nil {
		return nil, err
	}
	newToken := &auth.AuthToken{
		AccessToken:  accToken,
		RefreshToken: NewRefToken,
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(time.Minute * 15),
	}

	return newToken, nil
}
func (uc authUsecase) Logout(ctx context.Context, tokenID string) error {
	err := uc.refreshTokenRepo.Revoke(ctx, tokenID)
	if err != nil {
		return err
	}
	return nil
}
func (uc authUsecase) LogoutAll(ctx context.Context, userID string) error {
	err := uc.refreshTokenRepo.RevokeAllByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return nil
} // revokes all sessions across devices
// RequestPasswordReset — uses userRepo + resetTokenRepo + tokenService + notifService

// Password
// MASIH PERLU PERBAIKAN
func (uc authUsecase) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	if oldPassword == newPassword {
		return errors.New("new password must be different from old password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		return errors.New("hash failed")
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return auth.ErrUserNotFound
	}

	err = uc.userRepo.UpdatePasswordHash(ctx, userID, string(hash))
	if err != nil {
		return errors.New("password update failed")
	}

	err = uc.notifService.SendPasswordChanged(ctx, userID, user.Email)
	if err != nil {
		return errors.New("failed to send notification")
	}
	return nil
}

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

// PIN
func (uc *authUsecase) SetTradingPin(ctx context.Context, userID, pin string) error
func (uc *authUsecase) VerifyTradingPin(ctx context.Context, userID, pin string) (bool, error)
