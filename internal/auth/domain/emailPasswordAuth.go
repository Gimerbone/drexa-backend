package domain

import (
	"time"
)

type EmailPassword struct {
	Email             string
	PasswordHash      string
	PasswordUpdatedAt time.Time
}
