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
	// "drexa/internal/auth/repository"
	"drexa/internal/config"

	"gorm.io/gorm"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg *config.Config, db *gorm.DB) *Server {
	mux := http.NewServeMux()

	// TODO : Buka comment jika usecase sudah selesai di implement

	// Repositories
	// userRepo := repository.NewUserRepository(db)
	// authProviderRepo := repository.NewAuthProviderRepository(db)
	// refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	// resetTokenRepo := repository.NewPasswordResetTokenRepository(db)
	// kycRepo := repository.NewKycProfileRepository(db)

	// // Services
	// otpService := &auth.MockOTPService{}
	// notificationService := &auth.MockNotificationService{}
	// tokenService := auth.NewJWTTokenService(cfg.JWTSecret)

	// Usecases
	// authUsecase := usecase.NewAuthUsecase(userRepo, authProviderRepo, refreshTokenRepo, resetTokenRepo, otpService, notificationService, tokenService)
	// authProviderUsecase := usecase.NewAuthProviderUsecase(authProviderRepo, userRepo)
	// kycUsecase := usecase.NewKycUsecase(kycRepo, notificationService)
	// adminKycUsecase := usecase.NewAdminKycUsecase(kycRepo, notificationService)

	var authHandlers auth.AuthHandlers

	// Group by feature
	// authHandlers = auth.AuthHandlers{
	// 	Auth:         authUsecase,
	// 	AuthProvider: authProviderUsecase,
	// 	Kyc:          kycUsecase,
	// 	AdminKyc:     adminKycUsecase,
	// }

	addRoutes(
		mux,
		authHandlers,
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
