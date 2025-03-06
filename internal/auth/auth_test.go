package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	// Test cases
	tests := []struct {
		name        string
		userID      uuid.UUID
		tokenSecret string
		expiresIn   time.Duration
		wantErr     bool
	}{
		{
			name:        "valid token",
			userID:      uuid.New(),
			tokenSecret: "test-secret-key",
			expiresIn:   time.Hour,
			wantErr:     false,
		},
		{
			name:        "expired token",
			userID:      uuid.New(),
			tokenSecret: "test-secret-key",
			expiresIn:   -1 * time.Minute, // Already expired
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create token
			token, err := MakeJWT(tt.userID, tt.tokenSecret, tt.expiresIn)
			if err != nil {
				t.Fatalf("MakeJWT() error = %v", err)
			}

			// Validate token
			gotUserID, err := ValidateJWT(token, tt.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we don't expect an error, verify the user ID matches
			if !tt.wantErr && gotUserID != tt.userID {
				t.Errorf("ValidateJWT() gotUserID = %v, want %v", gotUserID, tt.userID)
			}
		})
	}
}

func TestValidateJWTWithWrongSecret(t *testing.T) {
	userID := uuid.New()
	correctSecret := "correct-secret"
	wrongSecret := "wrong-secret"
	expiresIn := time.Hour

	// Create token with correct secret
	token, err := MakeJWT(userID, correctSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	// Try to validate with wrong secret
	_, err = ValidateJWT(token, wrongSecret)
	if err == nil {
		t.Error("ValidateJWT() with wrong secret should return error, got nil")
	}
}

func TestValidateJWTWithInvalidToken(t *testing.T) {
	// Test with completely invalid token string
	invalidToken := "not-a-valid-jwt-token"
	secret := "some-secret"

	_, err := ValidateJWT(invalidToken, secret)
	if err == nil {
		t.Error("ValidateJWT() with invalid token should return error, got nil")
	}
}

func TestMakeJWTWithEmptySecret(t *testing.T) {
	userID := uuid.New()
	emptySecret := ""
	expiresIn := time.Hour

	// Create token with empty secret
	token, err := MakeJWT(userID, emptySecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT() with empty secret error = %v", err)
	}

	// Validate with empty secret
	gotUserID, err := ValidateJWT(token, emptySecret)
	if err != nil {
		t.Errorf("ValidateJWT() with empty secret error = %v", err)
	}

	if gotUserID != userID {
		t.Errorf("ValidateJWT() gotUserID = %v, want %v", gotUserID, userID)
	}
}
