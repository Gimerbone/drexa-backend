package auth

import (
	"context"
	"time"
)

// JWTClaims represents the payload embedded inside an access token
type JWTClaims struct {
	UserID          string    `json:"user_id"`
	Email           string    `json:"email"`
	IsEmailVerified bool      `json:"is_email_verified"`
	IsPhoneVerified bool      `json:"is_phone_verified"`
	IsKycVerified   bool      `json:"is_kyc_verified"` // lets middleware gate trading endpoints without a DB hit
	ExpiresAt       time.Time `json:"exp"`
	IssuedAt        time.Time `json:"iat"`
}

// OTPService abstracts OTP generation, storage, and delivery.
// Implement this for any provider — Twilio, AWS SNS, Firebase, or a mock for tests.
type OTPService interface {
	// GenerateAndSendEmail generates an OTP, stores it against the key, and sends it via email
	GenerateAndSendEmail(ctx context.Context, key, email string) error

	// GenerateAndSendSMS generates an OTP, stores it against the key, and sends it via SMS
	GenerateAndSendSMS(ctx context.Context, key, phone string) error

	// Verify checks the OTP for a given key — consumes it on success so it can't be reused
	Verify(ctx context.Context, key, otp string) (bool, error)
}

// TokenService abstracts JWT generation and validation.
// Implement this with golang-jwt/jwt or any compatible library.
type TokenService interface {
	// GenerateAccessToken issues a short-lived JWT containing user claims
	GenerateAccessToken(ctx context.Context, user *User) (string, error)

	// GenerateRefreshToken issues a long-lived opaque token for session renewal
	GenerateRefreshToken(ctx context.Context, userID string) (string, error)

	// ValidateAccessToken parses and validates a JWT, returns the embedded claims
	ValidateAccessToken(ctx context.Context, token string) (*JWTClaims, error)

	// HashToken hashes a raw token before storage — use for refresh and reset tokens
	HashToken(token string) string
}

// NotificationService abstracts user-facing notifications beyond OTP.
// Implement this for email/push once a provider is chosen — use MockNotificationService in the meantime.
type NotificationService interface {
	// SendKycApproved notifies the user their KYC was approved
	SendKycApproved(ctx context.Context, userID, email string) error

	// SendKycRejected notifies the user their KYC was rejected with a reason
	SendKycRejected(ctx context.Context, userID, email, reason string) error

	// SendPasswordChanged notifies the user their password was changed — security alert
	SendPasswordChanged(ctx context.Context, userID, email string) error

	// SendNewLogin notifies the user of a login from a new device — security alert
	SendNewLogin(ctx context.Context, userID, email, userAgent, ipAddress string) error

	// SendPasswordReset sends an email containing the raw reset token link
	SendPasswordReset(ctx context.Context, userID, email, rawToken string) error
}
