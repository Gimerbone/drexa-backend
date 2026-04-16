package service

import (
	"context"
	"drexa/internal/auth"
	"errors"
	"log"
	"sync"
	"time"
)

const (
	mockOTPValue = "1234"
	otpTTL       = 10 * time.Minute
)

type otpEntry struct {
	otp       string
	expiresAt time.Time
}

// mockOTPService is a test-only OTPService that always issues "1234".
// It stores OTPs in memory — no external provider needed.
//
// Replace with a real implementation (Twilio, AWS SNS, etc.) before going to production.
type mockOTPService struct {
	mu    sync.Mutex
	store map[string]otpEntry // key → OTP entry
}

// NewMockOTPService returns an OTPService suitable for local development and tests.
func NewMockOTPService() auth.OTPService {
	return &mockOTPService{
		store: make(map[string]otpEntry),
	}
}

// ── GenerateAndSendEmail ─────────────────────────────────────────────────────

func (s *mockOTPService) GenerateAndSendEmail(_ context.Context, key, email string) error {
	if key == "" || email == "" {
		return errors.New("mock_otp: key and email must not be empty")
	}

	s.set(key, mockOTPValue)

	// In a real implementation this would call an email provider.
	// Here we just log so test output stays visible.
	log.Printf("[mock_otp] EMAIL → %s | key=%s | otp=%s (expires in %s)",
		email, key, mockOTPValue, otpTTL)

	return nil
}

// ── GenerateAndSendSMS ───────────────────────────────────────────────────────

func (s *mockOTPService) GenerateAndSendSMS(_ context.Context, key, phone string) error {
	if key == "" || phone == "" {
		return errors.New("mock_otp: key and phone must not be empty")
	}

	s.set(key, mockOTPValue)

	// In a real implementation this would call Twilio / AWS SNS / etc.
	log.Printf("[mock_otp] SMS → %s | key=%s | otp=%s (expires in %s)",
		phone, key, mockOTPValue, otpTTL)

	return nil
}

// ── Verify ───────────────────────────────────────────────────────────────────

func (s *mockOTPService) Verify(_ context.Context, key, otp string) (bool, error) {
	if key == "" || otp == "" {
		return false, errors.New("mock_otp: key and otp must not be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.store[key]
	if !ok {
		return false, nil // no OTP issued for this key
	}

	if time.Now().After(entry.expiresAt) {
		delete(s.store, key) // clean up expired entry
		return false, nil
	}

	if entry.otp != otp {
		return false, nil // wrong code — do NOT consume so caller can handle retry logic
	}

	delete(s.store, key) // consume on success — prevents reuse
	return true, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func (s *mockOTPService) set(key, otp string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = otpEntry{
		otp:       otp,
		expiresAt: time.Now().Add(otpTTL),
	}
}
