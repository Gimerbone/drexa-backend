package domain

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthProvider interface {
	Authenticate() User
	GetProviderType() string
}

func NewEmailPasswordAuth(email, password string) (*EmailPassword, error) {
	emailpw := new(EmailPassword)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return emailpw, errors.New("something went wrong")
	}
	emailpw.Email = email
	emailpw.PasswordHash = string(hash)
	emailpw.PasswordUpdatedAt = time.Now()
	return emailpw, nil
}

func (acc EmailPassword) GetEmail() string {
	return acc.Email
}

func (acc *EmailPassword) UpdatePassword(pw string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), 10)
	if err != nil {
		return errors.New("failed to hash password")
	}
	acc.PasswordHash = string(hash)
	return nil
}
