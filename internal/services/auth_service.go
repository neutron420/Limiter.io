package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"limiter.io/internal/config"
	"limiter.io/internal/dto"
	"limiter.io/internal/mailer"
	"limiter.io/internal/models"
	"limiter.io/internal/repository"
	"limiter.io/internal/utils"

	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*models.User, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	// LoginMFA completes a login for MFA-enabled accounts (password + TOTP code).
	LoginMFA(ctx context.Context, req dto.MFALoginRequest) (*dto.AuthResponse, error)
	LoginWithGoogle(ctx context.Context, req dto.GoogleLoginRequest) (*dto.AuthResponse, error)
	Refresh(ctx context.Context, req dto.RefreshTokenRequest) (*dto.AuthResponse, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	ChangePassword(ctx context.Context, userID uuid.UUID, req dto.ChangePasswordRequest) error
	ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req dto.ResetPasswordRequest, clientIP string) error
}

type authService struct {
	userRepo   repository.UserRepository
	rtRepo     repository.RefreshTokenRepository
	subRepo    repository.SubscriptionRepository
	cacheRepo  repository.CacheRepository
	prtRepo    repository.PasswordResetTokenRepository
	mailer     mailer.Mailer
	cfg        *config.Config
}

func NewAuthService(
	userRepo repository.UserRepository,
	rtRepo repository.RefreshTokenRepository,
	subRepo repository.SubscriptionRepository,
	cacheRepo repository.CacheRepository,
	prtRepo repository.PasswordResetTokenRepository,
	mail mailer.Mailer,
	cfg *config.Config,
) AuthService {
	return &authService{
		userRepo:  userRepo,
		rtRepo:    rtRepo,
		subRepo:   subRepo,
		cacheRepo: cacheRepo,
		prtRepo:   prtRepo,
		mailer:    mail,
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

// issueSession creates the JWT + refresh token pair for an authenticated user.
func (s *authService) issueSession(ctx context.Context, user *models.User) (*dto.AuthResponse, error) {
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, s.cfg.JWTSecret, s.cfg.JWTAccessTTL)
	if err != nil {
		return nil, err
	}

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
		AvatarURL:    user.AvatarURL,
	}, nil
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	// MFA-enabled accounts must complete the TOTP step — no tokens yet.
	if user.MFAEnabled {
		return &dto.AuthResponse{MFARequired: true, UserEmail: user.Email}, nil
	}

	return s.issueSession(ctx, user)
}

// LoginMFA validates password + TOTP code and issues the session.
func (s *authService) LoginMFA(ctx context.Context, req dto.MFALoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}
	if !user.MFAEnabled || user.TOTPSecret == "" {
		return nil, errors.New("MFA is not enabled for this account")
	}
	if !ValidateTOTP(req.Code, user.TOTPSecret) {
		return nil, errors.New("invalid MFA code")
	}

	return s.issueSession(ctx, user)
}

func (s *authService) Refresh(ctx context.Context, req dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	rt, err := s.rtRepo.GetByToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if rt.Revoked {
		_ = s.rtRepo.RevokeByUserID(ctx, rt.UserID)
		return nil, errors.New("refresh token has been reused and revoked")
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, errors.New("refresh token expired")
	}

	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}

	_ = s.rtRepo.RevokeByUserID(ctx, user.ID)

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
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Silence error for security (prevent email enumeration)
		return nil
	}

	// Generate secure 6-digit numeric OTP code
	otpCode, err := utils.GenerateOTP(6)
	if err != nil {
		return err
	}

	// Hash the OTP code using SHA-256
	tokenHash := utils.HashAPIKey(otpCode)

	// Invalidate any existing password reset tokens for this user
	_ = s.prtRepo.DeleteByUserID(ctx, user.ID)

	// Save token in DB
	t := &models.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour expiry
	}
	if err := s.prtRepo.Create(ctx, t); err != nil {
		return err
	}

	// Build the reset link against the configured frontend origin.
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", strings.TrimRight(s.cfg.AppBaseURL, "/"), otpCode)

	// Send the reset email (Resend if configured, else logged by the mailer).
	subject := "Reset your Limiter.io password"
	html := fmt.Sprintf(`
		<div style="font-family:monospace;max-width:480px;margin:8px auto;border:2px solid #000;padding:24px;background:#fff;box-shadow:6px 6px 0px 0px #000;">
			<h2 style="text-transform:uppercase;letter-spacing:2px;margin-top:0;">Limiter.io — Security</h2>
			<p style="text-transform:uppercase;font-size:12px;color:#666;">Password Reset Request</p>
			<p>We received a request to reset your password. Use the verification code below to authorize this operation, or click the button:</p>
			
			<div style="background:#ea580c;color:#fff;font-size:32px;font-weight:bold;letter-spacing:6px;padding:16px;text-align:center;margin:20px 0;border:2px solid #000;box-shadow:4px 4px 0px 0px #000;">
				%s
			</div>

			<p style="text-align:center;margin:24px 0;">
				<a href="%s" style="display:inline-block;background:#000;color:#fff;padding:12px 20px;text-decoration:none;font-weight:bold;border:2px solid #000;text-transform:uppercase;letter-spacing:1px;">Click to Reset Password</a>
			</p>
			
			<p style="color:#888;font-size:10px;text-transform:uppercase;margin-top:24px;">This code and link expire in 1 hour. If you didn't request this, you can safely ignore this email.</p>
		</div>`, otpCode, resetURL)

	if err := s.mailer.Send(ctx, user.Email, subject, html); err != nil {
		// Don't fail the request (and don't leak existence) — log and also print the link for dev.
		log.Printf("Failed to send password reset email to %s: %v", user.Email, err)
		log.Printf("[dev fallback] reset OTP: %s | link: %s", otpCode, resetURL)
	}

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest, clientIP string) error {
	lockoutKey := fmt.Sprintf("lockout:otp:%s", clientIP)
	if val, err := s.cacheRepo.Get(ctx, lockoutKey); err == nil && val != "" {
		return errors.New("Too many failed attempts. OTP verification locked for 10 minutes.")
	}

	tokenHash := utils.HashAPIKey(req.Token)
	t, err := s.prtRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		// Log failed attempt and check limit
		attemptsKey := fmt.Sprintf("attempts:otp:%s", clientIP)
		attempts, errIncr := s.cacheRepo.Increment(ctx, attemptsKey, 10 * time.Minute)
		if errIncr == nil && attempts >= 3 {
			_ = s.cacheRepo.Set(ctx, lockoutKey, "locked", 10 * time.Minute)
			return errors.New("Too many failed attempts. OTP verification locked for 10 minutes.")
		}
		return errors.New("invalid or expired reset token")
	}

	user, err := s.userRepo.GetByID(ctx, t.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	newHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}

	user.PasswordHash = newHash
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Success: Delete lock/attempts trackers
	attemptsKey := fmt.Sprintf("attempts:otp:%s", clientIP)
	_ = s.cacheRepo.Delete(ctx, attemptsKey)
	_ = s.cacheRepo.Delete(ctx, lockoutKey)

	// Delete tokens after successful use
	_ = s.prtRepo.DeleteByUserID(ctx, user.ID)
	return nil
}

func (s *authService) LoginWithGoogle(ctx context.Context, req dto.GoogleLoginRequest) (*dto.AuthResponse, error) {
	info, err := utils.VerifyGoogleIDToken(req.IDToken)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByEmail(ctx, info.Email)
	if err != nil {
		// Create a new user since they don't exist yet
		// We generate a secure random password hash for the GORM schema constraint
		randPassword, err := utils.GenerateRandomToken(32)
		if err != nil {
			return nil, err
		}
		hashedPassword, err := utils.HashPassword(randPassword)
		if err != nil {
			return nil, err
		}

		user = &models.User{
			ID:           uuid.New(),
			Email:        info.Email,
			PasswordHash: hashedPassword,
			AvatarURL:    info.Picture,
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
			BillingMetadata: models.JSONMap{"source": "google_oauth"},
		}

		if err := s.subRepo.Create(ctx, sub); err != nil {
			return nil, err
		}
	} else {
		// Update profile picture if it has changed
		if user.AvatarURL != info.Picture {
			user.AvatarURL = info.Picture
			_ = s.userRepo.Update(ctx, user)
		}
	}

	// Generate standard JWT Access Token
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, s.cfg.JWTSecret, s.cfg.JWTAccessTTL)
	if err != nil {
		return nil, err
	}

	// Generate standard secure Refresh Token
	rawRefreshToken, err := utils.GenerateRandomToken(32)
	if err != nil {
		return nil, err
	}

	rt := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     rawRefreshToken,
		ExpiresAt: time.Now().Add(s.cfg.JWTRefreshTTL),
	}
	if err := s.rtRepo.Create(ctx, rt); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		UserEmail:    user.Email,
		UserID:       user.ID,
		AvatarURL:    user.AvatarURL,
	}, nil
}

