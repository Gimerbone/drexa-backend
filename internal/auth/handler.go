package auth

type AuthHandlers struct {
	Auth         AuthUsecase
	AuthProvider AuthProviderUsecase
	Kyc          KycUsecase
	AdminKyc     AdminKycUsecase
}

// TODO : Implement all handlers
