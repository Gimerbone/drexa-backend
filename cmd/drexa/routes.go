package main

import (
	"drexa/internal/auth"
	"net/http"
)

func addRoutes(
	mux *http.ServeMux,
	authUc auth.AuthUsecase,
	authProviderUC auth.AuthProviderUsecase,
	kycUc auth.KycUsecase,
	adminKycUc auth.AdminKycUsecase,
) {
	mux.Handle("/", http.NotFoundHandler())

	// TODO : IMPLEMENT ALL

	// Auth
	mux.Handle("POST /api/v1/auth/register", auth.HandleRegister(authUc))
	mux.Handle("POST /api/v1/auth/login", auth.HandleLogin(authUc))
	mux.Handle("POST /api/v1/auth/logout", auth.HandleLogout(authUc))
	mux.Handle("POST /api/v1/auth/refresh", auth.HandleRefreshToken(authUc))
	mux.Handle("POST /api/v1/auth/verify/email", auth.HandleVerifyEmail(authUc))
	mux.Handle("POST /api/v1/auth/verify/phone", auth.HandleVerifyPhone(authUc))
	mux.Handle("POST /api/v1/auth/password/reset", auth.HandleRequestPasswordReset(authUc))
	mux.Handle("POST /api/v1/auth/oauth/register", auth.HandleRegisterWithOAuth(authUc))
	mux.Handle("POST /api/v1/auth/oauth/login", auth.HandleLoginWithOAuth(authUc))

	mux.Handle("POST /api/v1/auth/pin/set", auth.HandleSetTradingPin(authUc))
	mux.Handle("POST /api/v1/auth/pin/verify", auth.HandleVerifyTradingPin(authUc))

	// Auth providers
	//mux.Handle("GET  /api/v1/auth/providers", auth.HandleGetAuthMethods(*authProviderUC))
	//mux.Handle("POST /api/v1/auth/providers/link", auth.HandleLinkAuthProvider(*authProviderUC))
	//mux.Handle("DELETE /api/v1/auth/providers/{id}", auth.HandleUnlinkAuthProvider(*authProviderUC))
	// KYC — user facing
	//mux.Handle("POST /api/v1/kyc/submit", auth.HandleKycSubmit(*kycUc))
	//mux.Handle("GET  /api/v1/kyc/status", auth.HandleKycStatus(*kycUc))

	// KYC — admin facing
	//mux.Handle("GET  /api/v1/admin/kyc", auth.HandleAdminKycList(*adminKycUc))
	//mux.Handle("GET  /api/v1/admin/kyc/{id}", auth.HandleAdminKycGet(*adminKycUc))
	//mux.Handle("POST /api/v1/admin/kyc/{id}/approve", auth.HandleAdminKycApprove(*adminKycUc))
	//mux.Handle("POST /api/v1/admin/kyc/{id}/reject", auth.HandleAdminKycReject(*adminKycUc))

}
