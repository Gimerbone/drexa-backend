package auth

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// ─── Entities ───────────────────────────────────────────────────────────────

type User struct {
	UserID          string         `gorm:"primaryKey;column:user_id"`
	UserName        string         `gorm:"column:username"`
	Email           string         `gorm:"column:email;uniqueIndex"`
	PasswordHash    string         `gorm:"column:password_hash"`
	PhoneNumber     string         `gorm:"column:phone_number"`
	TradingPinHash  string         `gorm:"column:trading_pin_hash"`
	IsEmailVerified bool           `gorm:"column:is_email_verified;default:false"`
	IsPhoneVerified bool           `gorm:"column:is_phone_verified;default:false"`
	LastLoginAt     time.Time      `gorm:"column:last_login_at"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime"`
	ModifiedAt      time.Time      `gorm:"column:modified_at;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"` // soft delete — required for OJK audit trail

	// Associations — not columns, loaded separately via Preload()
	KycProfile  KycProfile     `gorm:"foreignKey:UserID"`
	AuthMethods []AuthProvider `gorm:"foreignKey:UserID"`
}

type AuthProvider struct {
	AuthID      string    `gorm:"primaryKey;column:auth_id"`
	UserID      string    `gorm:"column:user_id;index"` // FK to users
	Provider    string    `gorm:"column:provider"`      // e.g. "google", "apple"
	ProviderUID string    `gorm:"column:provider_uid"`  // provider's unique user ID
	Email       string    `gorm:"column:email"`         // kept for auto-link logic on duplicate email
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

// KycStatus is a defined type to prevent arbitrary string values
type KycStatus string

const (
	KycStatusPending  KycStatus = "pending"
	KycStatusApproved KycStatus = "approved"
	KycStatusRejected KycStatus = "rejected"
	KycStatusExpired  KycStatus = "expired"
)

type KycAddress struct {
	Street    string `gorm:"column:address_street"`
	RTRW      string `gorm:"column:address_rt_rw"`
	Kelurahan string `gorm:"column:address_kelurahan"`
	Kecamatan string `gorm:"column:address_kecamatan"`
	Kabupaten string `gorm:"column:address_kabupaten"`
	Provinsi  string `gorm:"column:address_provinsi"`
	KodePos   string `gorm:"column:address_kode_pos"`
}

type KycProfile struct {
	KycID              string     `gorm:"primaryKey;column:kyc_id"`
	UserID             string     `gorm:"column:user_id;uniqueIndex"` // FK to users, one-to-one
	NIKEncrypted       string     `gorm:"column:nik_encrypted"`       // AES-256 encrypted, not hashed
	FullName           string     `gorm:"column:full_name"`
	DateOfBirth        time.Time  `gorm:"column:date_of_birth"`
	Address            KycAddress `gorm:"embedded"`                               // flattened into kyc_profiles table using column tags above
	DocumentImagePath  string     `gorm:"column:document_image_path"`             // path/URL to encrypted object storage
	FaceImagePath      string     `gorm:"column:face_image_path"`                 // path/URL to encrypted object storage
	VerificationSource string     `gorm:"column:verification_source"`             // e.g. "verihubs", "vida"
	DukcapilVerified   bool       `gorm:"column:dukcapil_verified;default:false"` // was NIK validated against Dukcapil?
	RejectionReason    string     `gorm:"column:rejection_reason"`                // populated if status = rejected
	SubmittedAt        time.Time  `gorm:"column:submitted_at"`
	VerifiedAt         *time.Time `gorm:"column:verified_at"` // pointer — nil until approved
	ExpiresAt          *time.Time `gorm:"column:expires_at"`  // KYC can expire and require re-verification
	Status             KycStatus  `gorm:"column:status;default:pending"`
	CreatedAt          time.Time  `gorm:"column:created_at;autoCreateTime"`
	ModifiedAt         time.Time  `gorm:"column:modified_at;autoUpdateTime"`
}

type AuthToken struct {
	AccessToken  string
	RefreshToken string
	TokenType    string    // always "Bearer"
	ExpiresAt    time.Time // when the AccessToken expires — lets client proactively refresh
	// note: AuthToken is never persisted — no GORM tags needed
	// AccessToken is a stateless JWT, RefreshToken is stored separately in refresh_tokens table
}

// RefreshToken represents a persisted refresh token for session management
type RefreshToken struct {
	TokenID   string     `gorm:"primaryKey;column:token_id"`
	UserID    string     `gorm:"column:user_id;index"`          // FK to users
	TokenHash string     `gorm:"column:token_hash;uniqueIndex"` // store hashed, never plaintext
	UserAgent string     `gorm:"column:user_agent"`             // which device/browser issued this token
	IPAddress string     `gorm:"column:ip_address"`             // for suspicious activity detection
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	RevokedAt *time.Time `gorm:"column:revoked_at"` // nil if still valid, set on logout
}

type PasswordResetToken struct {
	TokenID   string     `gorm:"primaryKey;column:token_id"`
	UserID    string     `gorm:"column:user_id;index"`          // FK to users
	TokenHash string     `gorm:"column:token_hash;uniqueIndex"` // hashed before storage
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at"` // nil until redeemed
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
}

// ─── Domain Errors ───────────────────────────────────────────────────────────

var (
	// User
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")

	// Auth provider
	ErrAuthProviderNotFound      = errors.New("auth provider not found")
	ErrAuthProviderAlreadyExists = errors.New("auth provider already linked")

	// Token
	ErrTokenInvalid = errors.New("token is invalid")
	ErrTokenExpired = errors.New("token has expired")

	// KYC
	ErrKycNotFound         = errors.New("kyc profile not found")
	ErrKycAlreadySubmitted = errors.New("kyc already submitted")
	ErrKycNotApproved      = errors.New("kyc not approved")

	// OTP
	ErrOTPInvalid = errors.New("otp is invalid")
	ErrOTPExpired = errors.New("otp has expired")
)
