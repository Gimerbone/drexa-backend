// cmd/drexa/server.go
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"drexa/internal/auth"
	authRepo "drexa/internal/auth/repository"
	authSvc "drexa/internal/auth/service"
	authUc "drexa/internal/auth/usecase"
	"drexa/internal/config"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Server struct {
	httpServer *http.Server
}

var (
	mockSecretKey = uuid.NewString()
)

func NewServer(cfg *config.Config, db *gorm.DB) *Server {
	mux := http.NewServeMux()

	// Repositories
	userRepo := authRepo.NewUserRepository(db)
	authProviderRepo := authRepo.NewAuthProviderRepository(db)
	refreshTokenRepo := authRepo.NewRefreshTokenRepository(db)
	resetTokenRepo := authRepo.NewPasswordResetTokenRepository(db)
	// kycRepo := authRepo.NewKycProfileRepository(db)

	// // // Services
	otpService := authSvc.NewMockOTPService()
	notificationService := authSvc.NewMockNotificationService()
	tokenService := authSvc.NewTokenService([]byte(mockSecretKey), "drexa.api", time.Minute*15, time.Hour*24*7)

	// Usecases
	var authProviderUsecase auth.AuthProviderUsecase
	var kycUsecase auth.KycUsecase
	var adminKycUsecase auth.AdminKycUsecase

	authUsecase := authUc.NewAuthUsecase(userRepo, authProviderRepo, refreshTokenRepo, resetTokenRepo, otpService, notificationService, tokenService)
	// authProviderUsecase := authUc.NewAuthProviderUsecase(authProviderRepo, userRepo)
	// kycUsecase := authUc.NewKycUsecase(kycRepo, notificationService)
	// adminKycUsecase := authUc.NewAdminKycUsecase(kycRepo, notificationService)

	// Group by feature
	// authHandlers = auth.AuthHandlers{
	// 	Auth:         authUsecase,
	// 	AuthProvider: authProviderUsecase,
	// 	Kyc:          kycUsecase,
	// 	AdminKyc:     adminKycUsecase,
	// }

	addRoutes(
		mux,
		authUsecase,
		authProviderUsecase,
		kycUsecase,
		adminKycUsecase,
	)

	var handler http.Handler = mux
	// handler = middleware.Logging(handler)
	// handler = middleware.CORS(handler)
	// handler = middleware.Auth(handler)

	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.App.Port,
			Handler:      handler,
			ReadTimeout:  cfg.App.ReadTimeout,
			WriteTimeout: cfg.App.WriteTimeout,
			IdleTimeout:  cfg.App.IdleTimeout,
		},
	}
}

func (s *Server) Start(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	go func() {
		log.Printf("server listening on %s\n", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("error shutting down server: %w", err)
	}

	log.Println("server stopped cleanly")
	return nil
}
