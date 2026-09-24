package auth

import (
	"atlas/internal/domain"
	appErr "atlas/internal/errors"
	tokenHelper "atlas/internal/token"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo                  *Repository
	db                    *sql.DB
	accessTokenGenerator  AccessTokenGenerator
	refreshTokenGenerator RefreshTokenGenerator
}

func NewService(repo *Repository, db *sql.DB, tokenGenerator AccessTokenGenerator, refreshTokenGenerator RefreshTokenGenerator) *Service {
	return &Service{
		repo:                  repo,
		db:                    db,
		accessTokenGenerator:  tokenGenerator,
		refreshTokenGenerator: refreshTokenGenerator,
	}
}

func (s *Service) Register(ctx context.Context, email, password, confirmPassword, firstName, lastName string) (*RegistrationResponse, error) {
	existingUser, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		slog.Error(fmt.Sprintf("failed to get user by email: %v", err))
		return nil, appErr.ErrUserNotFound
	}

	if existingUser != nil {
		slog.Error(fmt.Sprintf("user already exists: %s", email))
		return nil, appErr.ErrUserAlreadyExists
	}

	if password != confirmPassword {
		slog.Error(fmt.Sprintf("register: passwords do not match"))
		return nil, appErr.ErrPasswordMismatch
	}

	if err := ValidatePassword(password); err != nil {
		slog.Error(fmt.Sprintf("register: invalid password"))
		return nil, err
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to hash password: %v", err))
		return nil, err
	}

	pid := uuid.NewString()
	sid := uuid.NewString()

	result := s.refreshTokenGenerator.GenerateRefreshToken()

	accessToken, err := s.accessTokenGenerator.GenerateAccessToken(pid, sid)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to generate access token: %v", err))
		return nil, err
	}

	hashedRefreshToken := tokenHelper.HashToken(result.RefreshToken)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to begin transaction: %v", err))
		return nil, err
	}
	defer tx.Rollback()

	txRepo := s.repo.WithTX(tx)

	user, err := txRepo.CreateUser(ctx, email, hashedPassword, firstName, lastName, pid)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to create user: %v", err))
		return nil, err
	}

	session := &domain.Session{
		UserID:    user.ID,
		JTI:       result.JTI,
		SessionID: sid,
		TokenHash: hashedRefreshToken,
		ExpiresAt: result.ExpiresAt,
	}

	if err := txRepo.CreateSession(ctx, session); err != nil {
		slog.Error(fmt.Sprintf("failed to create session: %v", err))
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		slog.Error(fmt.Sprintf("failed to commit transaction: %v", err))
		return nil, err
	}

	return &RegistrationResponse{
		User: &UserResponse{
			PublicID:  user.PublicID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		},
		Tokens: Tokens{
			AccessToken:  accessToken,
			RefreshToken: result.RefreshToken,
		},
	}, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		slog.Error(fmt.Sprintf("login: failed to get user by email: %v", err))
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErr.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := ComparePasswords(user.PasswordHash, password); err != nil {
		slog.Error(fmt.Sprintf("login: invalid password"))
		return nil, appErr.ErrInvalidCredentials
	}

	sid := uuid.NewString()

	accessToken, err := s.accessTokenGenerator.GenerateAccessToken(user.PublicID, sid)
	if err != nil {
		slog.Error(fmt.Sprintf("login: failed to generate access token: %v", err))
		return nil, err
	}

	result := s.refreshTokenGenerator.GenerateRefreshToken()
	tokenHash := tokenHelper.HashToken(result.RefreshToken)

	if err := s.repo.UpdateSessionTokenHash(ctx, sid, tokenHash, result.JTI, result.ExpiresAt); err != nil {
		slog.Error(fmt.Sprintf("login: failed to update session token hash: %v", err))
		return nil, err
	}

	session := &domain.Session{
		UserID:    user.ID,
		JTI:       result.JTI,
		SessionID: sid,
		TokenHash: tokenHash,
		ExpiresAt: result.ExpiresAt,
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		slog.Error(fmt.Sprintf("login: failed to create session: %v", err))
		return nil, err
	}

	return &LoginResponse{
		User: &UserResponse{
			PublicID:  user.PublicID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		},
		Tokens: Tokens{
			AccessToken:  accessToken,
			RefreshToken: result.RefreshToken,
		},
	}, nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	if err := s.repo.RevokeSession(ctx, sessionID); err != nil {
		slog.Error(fmt.Sprintf("logout: failed to revoke session: %v", err))
		return err
	}
	return nil
}

func (s *Service) RefreshAccessToken(ctx context.Context, userPublicID, sessionID, refreshToken string) (*Tokens, error) {
	session, err := s.repo.GetSessionBySID(ctx, sessionID)
	if err != nil {
		slog.Error(fmt.Sprintf("refresh: failed to get session: %v", err))
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErr.ErrMissingSession
		}
		return nil, err
	}

	if session.TokenHash != tokenHelper.HashToken(refreshToken) {
		slog.Error(fmt.Sprintf("refresh: token reuse detected for sid %s, revoking session", sessionID))
		if err := s.repo.RevokeSession(ctx, sessionID); err != nil {
			slog.Error(fmt.Sprintf("refresh: failed to revoke session after reuse detection: %v", err))
		}
		return nil, appErr.ErrMissingSession
	}

	accessToken, err := s.accessTokenGenerator.GenerateAccessToken(userPublicID, sessionID)
	if err != nil {
		slog.Error(fmt.Sprintf("refresh: failed to generate access token: %v", err))
		return nil, err
	}

	result := s.refreshTokenGenerator.GenerateRefreshToken()

	tokenHash := tokenHelper.HashToken(result.RefreshToken)

	if err := s.repo.UpdateSessionTokenHash(ctx, sessionID, tokenHash, result.JTI, result.ExpiresAt); err != nil {
		slog.Error(fmt.Sprintf("refresh: failed to update session token hash: %v", err))
		return nil, err
	}

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

func (s *Service) FindSessionByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	tokenHash = tokenHelper.HashToken(tokenHash)
	return s.repo.FindSessionByTokenHash(ctx, tokenHash)
}

func (s *Service) IsSessionActive(ctx context.Context, sessionID string) (bool, error) {
	return s.repo.IsSessionActive(ctx, sessionID)
}
