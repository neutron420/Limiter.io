package services

import (
	"context"
	"errors"

	"limiter.io/internal/dto"
	"limiter.io/internal/repository"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
)

// SecurityService covers MFA (TOTP) enrollment and session/device management.
type SecurityService interface {
	// MFA
	SetupMFA(ctx context.Context, userID uuid.UUID) (*dto.MFASetupResponse, error)
	VerifyAndEnableMFA(ctx context.Context, userID uuid.UUID, code string) error
	DisableMFA(ctx context.Context, userID uuid.UUID, code string) error
	// Sessions
	ListSessions(ctx context.Context, userID uuid.UUID, currentToken string) ([]dto.SessionResponse, error)
	RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error
}

type securityService struct {
	userRepo repository.UserRepository
	rtRepo   repository.RefreshTokenRepository
}

func NewSecurityService(userRepo repository.UserRepository, rtRepo repository.RefreshTokenRepository) SecurityService {
	return &securityService{userRepo: userRepo, rtRepo: rtRepo}
}

// SetupMFA generates a TOTP secret and stores it (disabled until verified).
func (s *securityService) SetupMFA(ctx context.Context, userID uuid.UUID) (*dto.MFASetupResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if user.MFAEnabled {
		return nil, errors.New("MFA is already enabled — disable it first to re-enroll")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Limiter.io",
		AccountName: user.Email,
	})
	if err != nil {
		return nil, err
	}

	user.TOTPSecret = key.Secret()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return &dto.MFASetupResponse{
		Secret:     key.Secret(),
		OTPAuthURL: key.URL(),
	}, nil
}

// VerifyAndEnableMFA turns MFA on after the user proves they have the secret.
func (s *securityService) VerifyAndEnableMFA(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}
	if user.TOTPSecret == "" {
		return errors.New("no MFA enrollment in progress — call setup first")
	}
	if !totp.Validate(code, user.TOTPSecret) {
		return errors.New("invalid verification code")
	}
	user.MFAEnabled = true
	return s.userRepo.Update(ctx, user)
}

// DisableMFA requires a valid current code to turn MFA off.
func (s *securityService) DisableMFA(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}
	if !user.MFAEnabled {
		return errors.New("MFA is not enabled")
	}
	if !totp.Validate(code, user.TOTPSecret) {
		return errors.New("invalid verification code")
	}
	user.MFAEnabled = false
	user.TOTPSecret = ""
	return s.userRepo.Update(ctx, user)
}

// ValidateTOTP is used by the login flow (exported helper on the service).
func ValidateTOTP(code, secret string) bool {
	return totp.Validate(code, secret)
}

func (s *securityService) ListSessions(ctx context.Context, userID uuid.UUID, currentToken string) ([]dto.SessionResponse, error) {
	tokens, err := s.rtRepo.ListActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.SessionResponse, len(tokens))
	for i, t := range tokens {
		resp[i] = dto.SessionResponse{
			ID:        t.ID,
			UserAgent: t.UserAgent,
			ClientIP:  t.ClientIP,
			CreatedAt: t.CreatedAt,
			ExpiresAt: t.ExpiresAt,
			Current:   currentToken != "" && t.Token == currentToken,
		}
	}
	return resp, nil
}

func (s *securityService) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	return s.rtRepo.RevokeByID(ctx, userID, sessionID)
}
