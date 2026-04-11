package service

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

func Login() {}

// func EmailHandler(email, pw, confpw string) (model.User, error) {

// }

func GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%04d", rand.Intn(10000))
}

func GenerateOTPforgotpw() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func HasMX(email string) bool {
	part := strings.Split(email, "@")
	if len(part) != 2 {
		return false
	}

	domain := part[1]
	mx, err := net.LookupMX(domain)
	return err == nil && len(mx) > 0
}
