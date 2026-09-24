package auth

import "time"

var AtlasRefreshTokenCookieName = "atlas_refresh_token"

type RefreshTokenResponse struct {
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	JTI          string    `json:"jti"`
}

type AccessTokenGenerator interface {
	GenerateAccessToken(userID, sid string) (string, error)
}

type RefreshTokenGenerator interface {
	GenerateRefreshToken() *RefreshTokenResponse
}
