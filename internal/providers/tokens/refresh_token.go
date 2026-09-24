package tokens

import (
	"time"

	"github.com/google/uuid"
)

func GenerateRefreshToken() *RefreshTokenResponse {
	expiresAt := time.Now().Add(time.Hour * 24 * 30).UTC()
	token := "RT" + uuid.NewString()
	return &RefreshTokenResponse{
		RefreshToken: token,
		ExpiresAt:    expiresAt,
		JTI:          uuid.NewString(),
	}
}
