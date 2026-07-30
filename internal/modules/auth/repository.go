package auth

import (
	"atlas/internal/domain"
	appErr "atlas/internal/errors"
	"context"
	"database/sql"
	"log"
	"time"
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

func (r *Repository) IsSessionActive(ctx context.Context, sessionID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM sessions WHERE sid = $1 AND expires_at > NOW() AND revoked_at IS NULL)`

	var active bool
	err := r.db.
		QueryRowContext(ctx, query, sessionID).
		Scan(&active)
	if err != nil {
		return false, err
	}
	return active, nil
}

func (r *Repository) RevokeSession(ctx context.Context, sessionID string) error {
	query := `UPDATE sessions SET revoked_at = NOW() WHERE sid = $1`

	sqlResult, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return err
	}

	rowsAffected, err := sqlResult.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		log.Printf("revoke session: no session found for sid %s (already revoked or nonexistent)", sessionID)
	}

	return nil
}

func (r *Repository) UpdateSessionTokenHash(ctx context.Context, sessionID, tokenHash, jti string, expiresAt time.Time) error {
	query := `UPDATE sessions SET token_hash = $2, jti = $3, expires_at = $4 WHERE sid = $1 AND revoked_at IS NULL AND expires_at > NOW()`

	sqlResult, err := r.db.ExecContext(ctx, query, sessionID, tokenHash, jti, expiresAt)
	if err != nil {
		return err
	}

	rowsAffected, err := sqlResult.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return appErr.ErrMissingSession
	}

	return nil
}

func (r *Repository) GetSessionBySID(ctx context.Context, sessionID string) (*domain.Session, error) {
	query := `SELECT token_hash, jti FROM sessions WHERE sid = $1 AND revoked_at IS NULL`

	var session domain.Session
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(&session.TokenHash, &session.JTI)
	if err != nil {
		return nil, err
	}
	return &session, nil
}
