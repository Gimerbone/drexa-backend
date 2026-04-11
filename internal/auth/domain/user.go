package domain

import (
	"time"
)

type User struct {
	UserId      string
	UserName    string
	PhoneNumber string
	KycProfile  KycProfile
	Otps        []OtpChallenge
	AuthMethod  []AuthProvider
	CreatedAt   time.Time
	modifiedAt  time.Time
}
