package usecase

import (
	"context"
	"drexa/internal/auth"
	"errors"
	"time"

	"github.com/google/uuid"
)

type authProvUsecase struct {
	userRepo         auth.UserRepository
	authProviderRepo auth.AuthProviderRepository
}

func (provUC authProvUsecase) LinkAuthProvider(ctx context.Context, userID, provider, providerUID string) (*auth.AuthProvider, error) {
	user, err := provUC.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, auth.ErrUserNotFound
	}

	authmethod := &auth.AuthProvider{
		AuthID:      uuid.NewString(),
		UserID:      userID,
		Provider:    provider,
		ProviderUID: providerUID,
		Email:       user.Email,
		CreatedAt:   time.Now(),
	}

	if provUC.authProviderRepo.Create(ctx, authmethod) != nil {
		return nil, errors.New("failed to store in db")
	}

	user.AuthMethods = append(user.AuthMethods, *authmethod)
	return authmethod, nil
}

func (provUC authProvUsecase) UnlinkAuthProvider(ctx context.Context, userID, authID string) error {
	user, err := provUC.userRepo.FindByID(ctx, userID)
	if err != nil {
		return auth.ErrUserNotFound
	}
	if len(user.AuthMethods) == 1 {
		return errors.New("cant delete the only auth method")
	}
	filter := []auth.AuthProvider{}
	for _, p := range user.AuthMethods {
		if p.AuthID != authID {
			filter = append(filter, p)
		}
	}
	user.AuthMethods = filter
	return nil
} // should block if it's the only auth method

func (provUC authProvUsecase) GetAuthMethods(ctx context.Context, userID string) ([]auth.AuthProvider, error) {
	user, err := provUC.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, auth.ErrUserNotFound
	}

	return user.AuthMethods, nil
}

func (provUC authProvUsecase) FindByProvider(ctx context.Context, provider, providerUID string) (*auth.AuthProvider, error) {
	prov, err := provUC.authProviderRepo.FindByProvider(ctx, provider, providerUID)
	if err != nil {
		return nil, err
	}
	return prov, nil
}
