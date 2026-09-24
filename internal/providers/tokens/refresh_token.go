package tokens

import (
	"time"

	"github.com/google/uuid"
)

type RefreshTokenGen struct {
}

func NewRefreshTokenGen() *RefreshTokenGen {
	return &RefreshTokenGen{}
}

func (r *RefreshTokenGen) GenerateRefreshToken() *RefreshTokenResponse {
	expiresAt := time.Now().Add(time.Hour * 24 * 30).UTC()
	token := "RT" + uuid.NewString()
	return &RefreshTokenResponse{
		RefreshToken: token,
		ExpiresAt:    expiresAt,
		JTI:          uuid.NewString(),
	}
}
