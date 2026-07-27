package domain

import "time"

type Session struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	JTI       string    `db:"jti"`
	SessionID string    `db:"session_id"`
	TokenHash string    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
	RevokedAt time.Time `db:"revoked_at"`
	CreatedAt time.Time `db:"created_at"`
}
