package authprovider

import "os/user"

type AuthProvider interface {
	Authenticate() user.User
	GetProviderType() string
}
