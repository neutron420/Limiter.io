package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"limiter.io/internal/models"
	"gorm.io/gorm"
)

type PasskeyService struct {
	db      *gorm.DB
	session map[string]*RegistrationSession
}

type RegistrationSession struct {
	UserID      string
	Challenge   string
	CreatedAt   time.Time
}

func NewPasskeyService(db *gorm.DB) *PasskeyService {
	return &PasskeyService{
		db:      db,
		session: make(map[string]*RegistrationSession),
	}
}

func generateChallenge() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func (s *PasskeyService) BeginRegistration(userID string) (*RegistrationSession, error) {
	challenge := generateChallenge()
	session := &RegistrationSession{
		UserID:    userID,
		Challenge: challenge,
		CreatedAt: time.Now(),
	}
	s.session[userID] = session
	return session, nil
}

func (s *PasskeyService) CompleteRegistration(userID, credentialID, publicKey, attestationType, aaguid, nickname string) (*models.Passkey, error) {
	if _, ok := s.session[userID]; !ok {
		return nil, fmt.Errorf("no registration session found")
	}
	delete(s.session, userID)

	passkey := &models.Passkey{
		UserID:          userID,
		CredentialID:    credentialID,
		PublicKey:       publicKey,
		AttestationType: attestationType,
		AAGUID:          aaguid,
		Nickname:        nickname,
		SignCount:       0,
	}
	if err := s.db.Create(passkey).Error; err != nil {
		return nil, err
	}
	return passkey, nil
}

func (s *PasskeyService) BeginLogin() string {
	return generateChallenge()
}

func (s *PasskeyService) CompleteLogin(credentialID, signature string) (*models.User, error) {
	var passkey models.Passkey
	if err := s.db.Where("credential_id = ?", credentialID).First(&passkey).Error; err != nil {
		return nil, fmt.Errorf("passkey not found")
	}
	now := time.Now()
	passkey.LastUsedAt = &now
	passkey.SignCount++
	s.db.Save(&passkey)

	var user models.User
	if err := s.db.First(&user, "id = ?", passkey.UserID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *PasskeyService) ListPasskeys(userID string) ([]models.Passkey, error) {
	var passkeys []models.Passkey
	err := s.db.Where("user_id = ?", userID).Find(&passkeys).Error
	return passkeys, err
}

func (s *PasskeyService) DeletePasskey(id, userID string) error {
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Passkey{}).Error
}
