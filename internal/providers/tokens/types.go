package tokens

import "time"

const (
	AccessToken  = "access_token"
	RefreshToken = "refresh_token"
)

type RefreshTokenResponse struct {
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	JTI          string    `json:"jti"`
}
