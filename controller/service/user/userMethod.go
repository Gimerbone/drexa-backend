package user

import (
	"time"

	authprovider "drexa.com/controller/service/AuthProvider"
	kycprofile "drexa.com/controller/service/KycProfile"
	otpchallenge "drexa.com/controller/service/OtpChallenge"
)

type User struct {
	UserId      string
	UserName    string
	PhoneNumber string
	KycProfile  kycprofile.KycProfile
	Otps        []otpchallenge.OtpChallenge
	AuthMethod  []authprovider.AuthProvider
	CreatedAt   time.Time
	modifiedAt  time.Time
}

func (user User) NewUser() {

}
