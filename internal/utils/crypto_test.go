package utils

import (
	"strings"
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	pw := "s3cret-passw0rd"
	hash, err := HashPassword(pw)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == pw {
		t.Fatal("hash must not equal the plaintext password")
	}
	if !CheckPasswordHash(pw, hash) {
		t.Error("CheckPasswordHash should succeed for the correct password")
	}
	if CheckPasswordHash("wrong-password", hash) {
		t.Error("CheckPasswordHash should fail for an incorrect password")
	}
}

func TestGenerateAPIKey(t *testing.T) {
	full, hash, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey error: %v", err)
	}
	if !strings.HasPrefix(full, "rk_live_") {
		t.Errorf("expected key to start with rk_live_, got %q", full)
	}
	if hash == full {
		t.Error("stored hash must differ from the plaintext key")
	}
	// Hash must be deterministic and match HashAPIKey.
	if HashAPIKey(full) != hash {
		t.Error("HashAPIKey(full) should equal the hash returned by GenerateAPIKey")
	}
	// Two generated keys must be unique.
	full2, _, _ := GenerateAPIKey()
	if full == full2 {
		t.Error("two generated API keys should not collide")
	}
}

func TestHashAPIKeyStable(t *testing.T) {
	a := HashAPIKey("rk_live_abc")
	b := HashAPIKey("rk_live_abc")
	if a != b {
		t.Error("HashAPIKey must be deterministic for the same input")
	}
	if HashAPIKey("rk_live_abc") == HashAPIKey("rk_live_xyz") {
		t.Error("different inputs must produce different hashes")
	}
}
