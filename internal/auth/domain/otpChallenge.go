package domain

import (
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type OtpChallenge struct {
	OtpId        string
	Target       string
	purpose      string
	CodeHash     string
	ExpiresAt    time.Time
	AttemptsLeft int
}

// func GenerateOTP() string {
// 	rand.Seed(time.Now().UnixNano())
// 	return fmt.Sprintf("%04d", rand.Intn(10000))
// }

// func GenerateOTPforgotpw() string {
// 	rand.Seed(time.Now().UnixNano())
// 	return fmt.Sprintf("%06d", rand.Intn(1000000))
// }

// func HasMX(email string) bool {
// 	part := strings.Split(email, "@")
// 	if len(part) != 2 {
// 		return false
// 	}

// 	domain := part[1]
// 	mx, err := net.LookupMX(domain)
// 	return err == nil && len(mx) > 0
// }

func Generate(target string, purpose string) *OtpChallenge {
	hashotp, err := bcrypt.GenerateFromPassword([]byte(GenerateOTP()), 10)
	for err != nil {
		hashotp, err = bcrypt.GenerateFromPassword([]byte(GenerateOTP()), 10)
	}
	otp := new(OtpChallenge)
	otp.CodeHash = string(hashotp)
	otp.Target = target
	otp.purpose = purpose
	otp.AttemptsLeft = 3
	otp.ExpiresAt = time.Now().Add(5 * time.Minute)
	return otp
}

func (otp OtpChallenge) Verify(code string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(otp.CodeHash), []byte(code))
	if err != nil {
		return false
	}
	return true
}

func (otp OtpChallenge) IsExpired() bool {
	if time.Now().After(otp.ExpiresAt) {
		return false
	}
	return true
}

func (otp OtpChallenge) DecreaseAttempt() {
	otp.AttemptsLeft--
}

func GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%04d", rand.Intn(10000))
}
