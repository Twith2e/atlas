package middleware

import (
	"atlas/internal/providers/tokens"
	"context"
)

var UserPublicIDContextKey = "pid"
var SessionIDContextKey = "sid"
var RefreshTokenContextKey = "rt"
var AtlasRefreshTokenCookieName = "atlas_refresh_token"

type Signer interface {
	ValidateAccessToken(tokenString string) (*tokens.AccessTokenClaims, error)
	ValidateRefreshToken(tokenString string) (*tokens.RefreshTokenClaims, error)
}

type SessionChecker interface {
	IsSessionActive(ctx context.Context, sessionID string) (bool, error)
}
