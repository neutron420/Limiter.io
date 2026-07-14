package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateAPIKey() (string, string, error) {
	// Generate random 32 bytes for the API key secret
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	rawKey := hex.EncodeToString(bytes)
	prefix := "rk_live_"
	fullKey := fmt.Sprintf("%s%s", prefix, rawKey)

	// Hash using SHA-256 for DB search/validation
	hash := sha256.Sum256([]byte(fullKey))
	hashedKey := hex.EncodeToString(hash[:])

	return fullKey, hashedKey, nil
}

func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateOTP(length int) (string, error) {
	const digits = "0123456789"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i := 0; i < length; i++ {
		bytes[i] = digits[int(bytes[i])%len(digits)]
	}
	return string(bytes), nil
}

