package services

import (
	"context"
	"errors"
	"time"

	"limiter.io/internal/config"
	"limiter.io/internal/dto"
	"limiter.io/internal/models"
	"limiter.io/internal/repository"
	"limiter.io/internal/utils"

	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	Refresh(ctx context.Context, req dto.RefreshTokenRequest) (*dto.AuthResponse, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	ChangePassword(ctx context.Context, userID uuid.UUID, req dto.ChangePasswordRequest) error
	ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequest) error
}

type authService struct {
	userRepo   repository.UserRepository
	rtRepo     repository.RefreshTokenRepository
	subRepo    repository.SubscriptionRepository
	cacheRepo  repository.CacheRepository
	cfg        *config.Config
}

func NewAuthService(
	userRepo repository.UserRepository,
	rtRepo repository.RefreshTokenRepository,
	subRepo repository.SubscriptionRepository,
	cacheRepo repository.CacheRepository,
	cfg *config.Config,
) AuthService {
	return &authService{
		userRepo:  userRepo,
		rtRepo:    rtRepo,
		subRepo:   subRepo,
		cacheRepo: cacheRepo,
		cfg:       cfg,
	}
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error) {
	// Check if email already exists
	existing, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("email is already registered")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Create a default FREE subscription
	sub := &models.Subscription{
		ID:              uuid.New(),
		UserID:          user.ID,
		PlanID:          "free",
		Status:          "active",
		StartsAt:        time.Now(),
		BillingMetadata: models.JSONMap{"source": "registration"},
	}

	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT Access Token
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, s.cfg.JWTSecret, s.cfg.JWTAccessTTL)
	if err != nil {
		return nil, err
	}

	// Generate secure refresh token
	rawRefreshToken, err := utils.GenerateRandomString(32)
	if err != nil {
		return nil, err
	}

	refreshToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     rawRefreshToken,
		ExpiresAt: time.Now().Add(s.cfg.JWTRefreshTTL),
		Revoked:   false,
	}

	if err := s.rtRepo.Create(ctx, refreshToken); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		UserEmail:    user.Email,
		UserID:       user.ID,
	}, nil
}

func (s *authService) Refresh(ctx context.Context, req dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	rt, err := s.rtRepo.GetByToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if rt.Revoked {
		// Security Alert: Refresh Token Reuse!
		// Revoke all tokens for this user for security
		s.rtRepo.RevokeByUserID(ctx, rt.UserID)
		return nil, errors.New("refresh token has been reused and revoked")
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, errors.New("refresh token expired")
	}

	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}

	// Rotate refresh token
	s.rtRepo.RevokeByUserID(ctx, user.ID)

	newRawRefreshToken, err := utils.GenerateRandomString(32)
	if err != nil {
		return nil, err
	}

	newRefreshToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     newRawRefreshToken,
		ExpiresAt: time.Now().Add(s.cfg.JWTRefreshTTL),
		Revoked:   false,
	}

	if err := s.rtRepo.Create(ctx, newRefreshToken); err != nil {
		return nil, err
	}

	newAccessToken, err := utils.GenerateAccessToken(user.ID, user.Email, s.cfg.JWTSecret, s.cfg.JWTAccessTTL)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRawRefreshToken,
		UserEmail:    user.Email,
		UserID:       user.ID,
	}, nil
}

func (s *authService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.rtRepo.RevokeByUserID(ctx, userID)
}

func (s *authService) ChangePassword(ctx context.Context, userID uuid.UUID, req dto.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if !utils.CheckPasswordHash(req.OldPassword, user.PasswordHash) {
		return errors.New("incorrect old password")
	}

	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = newHash
	return s.userRepo.Update(ctx, user)
}

func (s *authService) ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequest) error {
	// Secure Architecture:
	// Verify user exists, generate token, send password reset link email (mock for backend-only platform).
	_, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Silence error for security (prevent email enumeration)
		return nil
	}

	// In real-world, push PasswordResetEvent to a queue (Kafka or mail processor)
	return nil
}
