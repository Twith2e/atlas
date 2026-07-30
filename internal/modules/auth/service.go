package auth

import (
	"atlas/internal/domain"
	appErr "atlas/internal/errors"
	tokenHelper "atlas/internal/token"
	"database/sql"
	"errors"
	"log"

	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo           *Repository
	db             *sql.DB
	tokenGenerator TokenGenerator
}

func NewService(repo *Repository, db *sql.DB, tokenGenerator TokenGenerator) *Service {
	return &Service{repo: repo, db: db, tokenGenerator: tokenGenerator}
}

func (s *Service) Register(ctx context.Context, email, password, confirmPassword, firstName, lastName string) (*RegistrationResponse, error) {
	existingUser, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("failed to get user by email: %v", err)
		return nil, err
	}

	if existingUser != nil {
		log.Printf("user already exists: %s", email)
		return nil, appErr.ErrUserAlreadyExists
	}

	if password != confirmPassword {
		log.Printf("register: passwords do not match")
		return nil, appErr.ErrPasswordMismatch
	}

	if err := ValidatePassword(password); err != nil {
		log.Printf("register: invalid password")
		return nil, err
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		log.Printf("failed to hash password: %v", err)
		return nil, err
	}

	pid := uuid.NewString()
	sid := uuid.NewString()

	refreshToken, jti, expiresAt, err := s.tokenGenerator.GenerateRefreshToken(pid, sid)
	if err != nil {
		log.Printf("failed to generate refresh token: %v", err)
		return nil, err
	}

	accessToken, err := s.tokenGenerator.GenerateAccessToken(pid, sid)
	if err != nil {
		log.Printf("failed to generate access token: %v", err)
		return nil, err
	}

	hashedRefreshToken := tokenHelper.HashToken(refreshToken)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("failed to begin transaction: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	txRepo := s.repo.WithTX(tx)

	user, err := txRepo.CreateUser(ctx, email, hashedPassword, firstName, lastName, pid)
	if err != nil {
		log.Printf("failed to create user: %v", err)
		return nil, err
	}

	session := &domain.Session{
		UserID:    user.ID,
		JTI:       jti,
		SessionID: sid,
		TokenHash: hashedRefreshToken,
		ExpiresAt: expiresAt,
	}

	if err := txRepo.CreateSession(ctx, session); err != nil {
		log.Printf("failed to create session: %v", err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("failed to commit transaction: %v", err)
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
			RefreshToken: refreshToken,
		},
	}, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		log.Printf("login: failed to get user by email: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErr.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := ComparePasswords(user.PasswordHash, password); err != nil {
		log.Printf("login: invalid password")
		return nil, appErr.ErrInvalidCredentials
	}

	sid := uuid.NewString()

	accessToken, err := s.tokenGenerator.GenerateAccessToken(user.PublicID, sid)
	if err != nil {
		log.Printf("login: failed to generate access token: %v", err)
		return nil, err
	}

	refreshToken, jti, expiresAt, err := s.tokenGenerator.GenerateRefreshToken(user.PublicID, sid)
	if err != nil {
		log.Printf("login: failed to generate refresh token: %v", err)
		return nil, err
	}

	tokenHash := tokenHelper.HashToken(refreshToken)

	session := &domain.Session{
		UserID:    user.ID,
		JTI:       jti,
		SessionID: sid,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		log.Printf("login: failed to create session: %v", err)
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
			RefreshToken: refreshToken,
		},
	}, nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	if err := s.repo.RevokeSession(ctx, sessionID); err != nil {
		log.Printf("logout: failed to revoke session: %v", err)
		return err
	}
	return nil
}

func (s *Service) RefreshAccessToken(ctx context.Context, userPublicID, sessionID, refreshToken string) (*Tokens, error) {
	session, err := s.repo.GetSessionBySID(ctx, sessionID)
	if err != nil {
		log.Printf("refresh: failed to get session: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErr.ErrMissingSession
		}
		return nil, err
	}

	if session.TokenHash != tokenHelper.HashToken(refreshToken) {
		log.Printf("refresh: token reuse detected for sid %s, revoking session", sessionID)
		if err := s.repo.RevokeSession(ctx, sessionID); err != nil {
			log.Printf("refresh: failed to revoke session after reuse detection: %v", err)
		}
		return nil, appErr.ErrMissingSession
	}

	accessToken, err := s.tokenGenerator.GenerateAccessToken(userPublicID, sessionID)
	if err != nil {
		log.Printf("refresh: failed to generate access token: %v", err)
		return nil, err
	}

	newRefreshToken, jti, expiresAt, err := s.tokenGenerator.GenerateRefreshToken(userPublicID, sessionID)
	if err != nil {
		log.Printf("refresh: failed to generate refresh token: %v", err)
		return nil, err
	}

	tokenHash := tokenHelper.HashToken(newRefreshToken)

	if err := s.repo.UpdateSessionTokenHash(ctx, sessionID, tokenHash, jti, expiresAt); err != nil {
		log.Printf("refresh: failed to update session token hash: %v", err)
		return nil, err
	}

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *Service) IsSessionActive(ctx context.Context, sessionID string) (bool, error) {
	return s.repo.IsSessionActive(ctx, sessionID)
}
