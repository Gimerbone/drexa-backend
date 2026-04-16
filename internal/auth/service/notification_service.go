package service

import (
	"context"
	"drexa/internal/auth"
	"log"
)

// mockNotificationService is a test-only NotificationService that logs all
// notifications to stdout instead of calling a real provider (email/push/etc.).
//
// Replace with a real implementation before going to production.
type mockNotificationService struct{}

// NewMockNotificationService returns a NotificationService suitable for local development and tests.
func NewMockNotificationService() auth.NotificationService {
	return &mockNotificationService{}
}

func (s *mockNotificationService) SendKycApproved(_ context.Context, userID, email string) error {
	log.Printf("[mock_notification] KYC_APPROVED → userID=%s email=%s", userID, email)
	return nil
}

func (s *mockNotificationService) SendKycRejected(_ context.Context, userID, email, reason string) error {
	log.Printf("[mock_notification] KYC_REJECTED → userID=%s email=%s reason=%q", userID, email, reason)
	return nil
}

func (s *mockNotificationService) SendPasswordChanged(_ context.Context, userID, email string) error {
	log.Printf("[mock_notification] PASSWORD_CHANGED → userID=%s email=%s", userID, email)
	return nil
}

func (s *mockNotificationService) SendNewLogin(_ context.Context, userID, email, userAgent, ipAddress string) error {
	log.Printf("[mock_notification] NEW_LOGIN → userID=%s email=%s ip=%s ua=%q", userID, email, ipAddress, userAgent)
	return nil
}

func (s *mockNotificationService) SendPasswordReset(_ context.Context, userID, email, rawToken string) error {
	log.Printf("[mock_notification] PASSWORD_RESET → userID=%s email=%s token=%s", userID, email, rawToken)
	return nil
}
