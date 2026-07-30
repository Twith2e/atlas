package auth

import (
	appErr "atlas/internal/errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pw string) (string, error) {
	hashedPasswordByte, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPasswordByte), nil
}

func ValidatePassword(pw string) error {
	if len(pw) < 8 {
		return appErr.ErrInvalidPasswordLength
	}

	var hasUpper, hasLower, hasSpecialChar, hasNum bool

	for _, r := range pw {
		if r >= 'A' && r <= 'Z' {
			hasUpper = true
		}
		if r >= 'a' && r <= 'z' {
			hasLower = true
		}
		if r >= '0' && r <= '9' {
			hasNum = true
		}
		if r == '!' || r == '@' || r == '?' || r == '$' || r == '%' || r == '#' || r == '^' || r == '*' {
			hasSpecialChar = true
		}

		if hasUpper && hasLower && hasNum && hasSpecialChar {
			return nil
		}
	}

	return appErr.ErrInvalidPassword
}

func ComparePasswords(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return appErr.ErrPasswordMismatch
	}

	return nil
}

func setRefreshCookie(c *gin.Context, token, env string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     AtlasRefreshTokenCookieName,
		Value:    token,
		Path:     "/api/v1/auth",
		Expires:  time.Now().Add(time.Hour * 24 * 30).UTC(),
		SameSite: http.SameSiteStrictMode,
		Secure:   env == "production",
		HttpOnly: true,
	})
}

func clearRefreshCookie(c *gin.Context, env string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     AtlasRefreshTokenCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		Secure:   env == "production",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		HttpOnly: true,
	})
}
