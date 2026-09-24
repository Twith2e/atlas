package middleware

import (
	"atlas/internal/domain"
	"atlas/internal/providers/tokens"
	"context"
)

var UserPublicIDContextKey = "pid"
var SessionIDContextKey = "sid"
var RefreshTokenContextKey = "rt"
var AtlasRefreshTokenCookieName = "atlas_refresh_token"

type Signer interface {
	ValidateAccessToken(tokenString string) (*tokens.AccessTokenClaims, error)
}

type SessionChecker interface {
	FindSessionByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)
	IsSessionActive(ctx context.Context, sessionID string) (bool, error)
}
