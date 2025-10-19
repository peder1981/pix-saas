package security

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewJWTService(t *testing.T) {
	secretKey := []byte("test-secret-key")
	accessTTL := 15 * time.Minute
	refreshTTL := 7 * 24 * time.Hour

	service := NewJWTService(secretKey, accessTTL, refreshTTL)
	if service == nil {
		t.Fatal("NewJWTService() returned nil")
	}
}

func TestGenerateAccessToken(t *testing.T) {
	service := NewJWTService([]byte("test-secret"), 15*time.Minute, 7*24*time.Hour)

	userID := uuid.New()
	merchantID := uuid.New()
	email := "test@example.com"
	role := "merchant"

	token, expiresAt, err := service.GenerateAccessToken(userID, &merchantID, email, role)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateAccessToken() returned empty token")
	}

	if expiresAt.Before(time.Now()) {
		t.Error("GenerateAccessToken() expiresAt is in the past")
	}

	if expiresAt.After(time.Now().Add(20 * time.Minute)) {
		t.Error("GenerateAccessToken() expiresAt is too far in the future")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	service := NewJWTService([]byte("test-secret"), 15*time.Minute, 7*24*time.Hour)

	userID := uuid.New()

	token, err := service.GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateRefreshToken() returned empty token")
	}
}

func TestValidateAccessToken(t *testing.T) {
	service := NewJWTService([]byte("test-secret"), 15*time.Minute, 7*24*time.Hour)

	userID := uuid.New()
	merchantID := uuid.New()
	email := "test@example.com"
	role := "merchant"

	// Generate token
	token, _, err := service.GenerateAccessToken(userID, &merchantID, email, role)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	// Validate token
	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("ValidateToken() userID = %v, want %v", claims.UserID, userID)
	}

	if claims.MerchantID == nil || *claims.MerchantID != merchantID {
		t.Errorf("ValidateToken() merchantID = %v, want %v", claims.MerchantID, merchantID)
	}

	if claims.Role != role {
		t.Errorf("ValidateToken() role = %v, want %v", claims.Role, role)
	}
}

func TestValidateRefreshToken(t *testing.T) {
	service := NewJWTService([]byte("test-secret"), 15*time.Minute, 7*24*time.Hour)

	userID := uuid.New()

	// Generate token
	token, err := service.GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	// Validate token
	validatedUserID, err := service.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("ValidateRefreshToken() error = %v", err)
	}

	if validatedUserID != userID {
		t.Errorf("ValidateRefreshToken() userID = %v, want %v", validatedUserID, userID)
	}
}

func TestValidateInvalidToken(t *testing.T) {
	service := NewJWTService([]byte("test-secret"), 15*time.Minute, 7*24*time.Hour)

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "invalid format",
			token: "invalid.token.format",
		},
		{
			name:  "malformed token",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ValidateToken(tt.token)
			if err == nil {
				t.Error("ValidateToken() expected error, got nil")
			}
		})
	}
}

func TestValidateTokenWithWrongSecret(t *testing.T) {
	service1 := NewJWTService([]byte("secret1"), 15*time.Minute, 7*24*time.Hour)
	service2 := NewJWTService([]byte("secret2"), 15*time.Minute, 7*24*time.Hour)

	userID := uuid.New()
	merchantID := uuid.New()
	email := "test@example.com"

	// Generate with service1
	token, _, err := service1.GenerateAccessToken(userID, &merchantID, email, "merchant")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	// Try to validate with service2 (different secret)
	_, err = service2.ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken() with wrong secret should fail")
	}
}

func TestExpiredToken(t *testing.T) {
	// Create service with very short TTL
	service := NewJWTService([]byte("test-secret"), 1*time.Millisecond, 1*time.Millisecond)

	userID := uuid.New()
	merchantID := uuid.New()
	email := "test@example.com"

	token, _, err := service.GenerateAccessToken(userID, &merchantID, email, "merchant")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	_, err = service.ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken() should fail for expired token")
	}
}
