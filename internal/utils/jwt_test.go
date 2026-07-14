package utils

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAccessTokenRoundTrip(t *testing.T) {
	secret := "test-secret"
	uid := uuid.New()
	email := "dev@limiter.io"

	token, err := GenerateAccessToken(uid, email, secret, time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken error: %v", err)
	}

	claims, err := ValidateAccessToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateAccessToken error: %v", err)
	}
	if claims.UserID != uid {
		t.Errorf("UserID mismatch: got %v want %v", claims.UserID, uid)
	}
	if claims.Email != email {
		t.Errorf("Email mismatch: got %q want %q", claims.Email, email)
	}
}

func TestValidateRejectsWrongSecret(t *testing.T) {
	token, _ := GenerateAccessToken(uuid.New(), "a@b.co", "right-secret", time.Minute)
	if _, err := ValidateAccessToken(token, "wrong-secret"); err == nil {
		t.Error("expected error validating token with the wrong secret")
	}
}

func TestValidateRejectsExpiredToken(t *testing.T) {
	secret := "s"
	token, _ := GenerateAccessToken(uuid.New(), "a@b.co", secret, -time.Minute) // already expired
	if _, err := ValidateAccessToken(token, secret); err == nil {
		t.Error("expected error validating an expired token")
	}
}

func TestGenerateRandomString(t *testing.T) {
	s, err := GenerateRandomString(16)
	if err != nil {
		t.Fatalf("GenerateRandomString error: %v", err)
	}
	if len(s) != 32 { // hex encoding doubles the byte length
		t.Errorf("expected 32 hex chars for 16 bytes, got %d", len(s))
	}
	s2, _ := GenerateRandomString(16)
	if s == s2 {
		t.Error("two random strings should not be equal")
	}
}
