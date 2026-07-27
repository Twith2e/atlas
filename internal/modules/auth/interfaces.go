package auth

import "time"

type TokenGenerator interface {
	GenerateAccessToken(userID, sid string) (string, error)
	GenerateRefreshToken(userID, sid string) (string, string, time.Time, error)
}
