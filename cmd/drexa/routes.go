package main

import (
	"drexa/internal/auth"
	"net/http"
)

func addRoutes(
	mux *http.ServeMux,
	authHandler auth.AuthHandlers,
) {
	mux.Handle("/", http.NotFoundHandler())

	// TODO : IMPLEMENT ALL

	// Auth
	// mux.Handle("POST /api/v1/auth/register", auth.HandleRegister(auth.Auth))
	// mux.Handle("POST /api/v1/auth/login", auth.HandleLogin(auth.Auth))
	// mux.Handle("POST /api/v1/auth/logout", auth.HandleLogout(auth.Auth))
	// mux.Handle("POST /api/v1/auth/refresh", auth.HandleRefreshToken(auth.Auth))
	// mux.Handle("POST /api/v1/auth/verify/email", auth.HandleVerifyEmail(auth.Auth))
	// mux.Handle("POST /api/v1/auth/verify/phone", auth.HandleVerifyPhone(auth.Auth))
	// mux.Handle("POST /api/v1/auth/password/reset", auth.HandleRequestPasswordReset(auth.Auth))
	// mux.Handle("POST /api/v1/auth/oauth/register", auth.HandleRegisterWithOAuth(auth.Auth))
	// mux.Handle("POST /api/v1/auth/oauth/login", auth.HandleLoginWithOAuth(auth.Auth))

	// // Auth providers
	// mux.Handle("GET  /api/v1/auth/providers", auth.HandleGetAuthMethods(auth.AuthProvider))
	// mux.Handle("POST /api/v1/auth/providers/link", auth.HandleLinkAuthProvider(auth.AuthProvider))
	// mux.Handle("DELETE /api/v1/auth/providers/{id}", auth.HandleUnlinkAuthProvider(auth.AuthProvider))

	// // KYC — user facing
	// mux.Handle("POST /api/v1/kyc/submit", auth.HandleKycSubmit(auth.Kyc))
	// mux.Handle("GET  /api/v1/kyc/status", auth.HandleKycStatus(auth.Kyc))

	// // KYC — admin facing
	// mux.Handle("GET  /api/v1/admin/kyc", auth.HandleAdminKycList(auth.AdminKyc))
	// mux.Handle("GET  /api/v1/admin/kyc/{id}", auth.HandleAdminKycGet(auth.AdminKyc))
	// mux.Handle("POST /api/v1/admin/kyc/{id}/approve", auth.HandleAdminKycApprove(auth.AdminKyc))
	// mux.Handle("POST /api/v1/admin/kyc/{id}/reject", auth.HandleAdminKycReject(auth.AdminKyc))
}
