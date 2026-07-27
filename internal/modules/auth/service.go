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

	hashedPassword, err := HashPassword(password)
	if err != nil {
		log.Printf("failed to hash password: %v", err)
		return nil, err
	}

	if password != confirmPassword {
		log.Printf("password mismatch: %s != %s", password, confirmPassword)
		return nil, appErr.ErrPasswordMismatch
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
