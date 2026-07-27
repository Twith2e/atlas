package auth

import (
	"atlas/internal/domain"
	"context"
	"database/sql"
)

type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

type Repository struct {
	db DBTX
}

func NewRepository(db DBTX) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithTX(tx *sql.Tx) *Repository {
	return &Repository{db: tx}
}

func (r *Repository) CreateUser(ctx context.Context, email, passwordHash, firstName, lastName, pid string) (*domain.User, error) {
	query := `INSERT INTO users (public_id, first_name, last_name, email, password_hash)
 			  VALUES ($1, $2, $3, $4, $5)
              RETURNING id, public_id, first_name, last_name, email, password_hash, created_at`

	var user domain.User
	err := r.db.
		QueryRowContext(ctx, query, pid, firstName, lastName, email, passwordHash).
		Scan(&user.ID, &user.PublicID, &user.FirstName, &user.LastName, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByPublicID(ctx context.Context, pid string) (*domain.User, error) {
	query := `SELECT id, public_id, first_name, last_name, email, password_hash, created_at FROM users WHERE public_id = $1`

	var user domain.User
	err := r.db.
		QueryRowContext(ctx, query, pid).
		Scan(&user.ID, &user.PublicID, &user.FirstName, &user.LastName, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, public_id, first_name, last_name, email, password_hash, created_at FROM users WHERE email = $1`

	var user domain.User
	err := r.db.
		QueryRowContext(ctx, query, email).
		Scan(&user.ID, &user.PublicID, &user.FirstName, &user.LastName, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) CreateSession(ctx context.Context, session *domain.Session) error {
	query := `INSERT INTO sessions (user_id, jti, sid, token_hash, expires_at, revoked_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, session.UserID, session.JTI, session.SessionID, session.TokenHash, session.ExpiresAt, session.RevokedAt)
	return err
}

func (r *Repository) GetSessionByTokenHash(ctx context.Context, id string) (*domain.Session, error) {
	query := `SELECT id, user_id, jti, session_id, token_hash, expires_at, revoked_at, created_at FROM sessions WHERE token_hash = $1`
	var session domain.Session
	err := r.db.
		QueryRowContext(ctx, query, id).
		Scan(&session.ID, &session.UserID, &session.JTI, &session.SessionID, &session.TokenHash, &session.ExpiresAt, &session.RevokedAt, &session.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &session, nil
}
