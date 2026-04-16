package usecase

import (
	"context"
	"drexa/internal/auth"
	"errors"
	"time"
)

type KycUseCase struct {
	userRepo         auth.UserRepository
	authProviderRepo auth.AuthProviderRepository
	kycRepo          auth.KycProfileRepository
}

func (uc KycUseCase) Submit(ctx context.Context, userID string, kyc *auth.KycProfile) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return auth.ErrUserNotFound
	}

	if uc.kycRepo.Create(ctx, kyc) != nil {
		return errors.New("failed storing kyc")
	}
	user.KycProfile = *kyc
	return nil
}

func (uc KycUseCase) GetByUserID(ctx context.Context, userID string) (*auth.KycProfile, error) {
	kyc, err := uc.kycRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("kyc cant found")
	}

	return kyc, nil
} // user checks their own status

func (uc KycUseCase) IsVerified(ctx context.Context, userID string) (bool, error)
func (uc KycUseCase) IsExpired(ctx context.Context, userID string) (bool, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, auth.ErrUserNotFound
	}

	if user.KycProfile.ExpiresAt.After(time.Now()) {
		return false, errors.New("kyc expired")
	}

	return true, nil
}
