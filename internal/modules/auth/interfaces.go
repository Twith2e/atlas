package auth

import "time"

var AtlasRefreshTokenCookieName = "atlas_refresh_token"

type TokenGenerator interface {
	GenerateAccessToken(userID, sid string) (string, error)
	GenerateRefreshToken(userID, sid string) (refreshToken, jti string, expiresAt time.Time, err error)
}
