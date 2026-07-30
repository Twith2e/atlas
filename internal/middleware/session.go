package middleware

import (
	appErr "atlas/internal/errors"
	"atlas/internal/response"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SessionGuard(signer Signer) gin.HandlerFunc {
	return func(c *gin.Context) {
		refreshToken, err := c.Cookie(AtlasRefreshTokenCookieName)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				log.Printf("session: no refresh token cookie: %v", err)
				mapped := response.MapError(appErr.ErrUnauthorized)
				c.AbortWithStatusJSON(http.StatusUnauthorized, response.ErrorResponse{
					Status: "error",
					Error:  &mapped.Error,
				})
				return
			}
			log.Printf("session: failed to get refresh token cookie: %v", err)
			mapped := response.MapError(err)
			c.AbortWithStatusJSON(mapped.Status, response.ErrorResponse{
				Status: "error",
				Error:  &mapped.Error,
			})
			return
		}

		claims, err := signer.ValidateRefreshToken(refreshToken)
		if err != nil {
			log.Printf("session: failed to validate refresh token: %v", err)
			mapped := response.MapError(appErr.ErrUnauthorized)
			c.AbortWithStatusJSON(mapped.Status, response.ErrorResponse{
				Status: "error",
				Error:  &mapped.Error,
			})
			return
		}

		c.Set(SessionIDContextKey, claims.SID)
		c.Set(RefreshTokenContextKey, refreshToken)
		c.Set(UserPublicIDContextKey, claims.Subject)

		c.Next()
	}
}

func RequireActiveSession(checker SessionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := c.GetString(SessionIDContextKey)

		active, err := checker.IsSessionActive(c, sid)
		if err != nil {
			log.Printf("refresh: failed to check session active: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, response.ErrorResponse{
				Status: "error",
				Error:  &response.APIError{Code: "internal_server_error", Message: "internal server error"},
			})
			return
		}

		if !active {
			mapped := response.MapError(appErr.ErrUnauthorized)
			c.AbortWithStatusJSON(mapped.Status, response.ErrorResponse{
				Status: "error",
				Error:  &mapped.Error,
			})
			return
		}

		c.Next()
	}
}
