package auth

import (
	"atlas/internal/providers/tokens"
)

var AtlasRefreshTokenCookieName = "atlas_refresh_token"

type AccessTokenGenerator interface {
	GenerateAccessToken(userID, sid string) (string, error)
}

type RefreshTokenGenerator interface {
	GenerateRefreshToken() *tokens.RefreshTokenResponse
}
