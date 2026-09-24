package tokens

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessTokenClaims struct {
	SID  string `json:"sid"`
	Type string `json:"typ"`
	jwt.RegisteredClaims
}

type JWT struct {
	accessSecret string
}

func NewJWT(accessSecret string) *JWT {
	return &JWT{
		accessSecret: accessSecret,
	}
}

func (j *JWT) GenerateAccessToken(userID, sid string) (string, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(time.Minute * 15)

	claims := AccessTokenClaims{
		SID:  sid,
		Type: AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.accessSecret))
}

func (j *JWT) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&AccessTokenClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing methodL %v", token.Header["alg"])
			}
			return []byte(j.accessSecret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if claims.Type != "access_token" {
		return nil, fmt.Errorf("invalid token type")
	}

	return claims, nil
}
