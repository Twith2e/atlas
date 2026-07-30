package middleware

import (
	appErr "atlas/internal/errors"
	"atlas/internal/response"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(signer Signer, sessionChecker SessionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.Request.Header.Get("Authorization")

		if authorization == "" {
			mapped := response.MapError(appErr.ErrUnauthorized)
			c.AbortWithStatusJSON(mapped.Status, response.ErrorResponse{
				Status: "error",
				Error:  &mapped.Error,
			})
			return
		}

		parts := strings.Fields(authorization)
		if len(parts) != 2 || parts[0] != "Bearer" {
			mapped := response.MapError(appErr.ErrUnauthorized)
			c.AbortWithStatusJSON(mapped.Status, response.ErrorResponse{
				Status: "error",
				Error:  &mapped.Error,
			})
			return
		}

		token := parts[1]

		if token == "" {
			mapped := response.MapError(appErr.ErrUnauthorized)
			c.AbortWithStatusJSON(mapped.Status, response.ErrorResponse{
				Status: "error",
				Error:  &mapped.Error,
			})
			return
		}

		claims, err := signer.ValidateAccessToken(token)
		if err != nil {
			mapped := response.MapError(appErr.ErrUnauthorized)
			c.AbortWithStatusJSON(mapped.Status, response.ErrorResponse{
				Status: "error",
				Error:  &mapped.Error,
			})
			return
		}

		sessionActive, err := sessionChecker.IsSessionActive(c, claims.SID)
		if err != nil {
			mapped := response.MapError(err)
			c.AbortWithStatusJSON(mapped.Status, response.ErrorResponse{
				Status: "error",
				Error:  &mapped.Error,
			})
			return
		}

		if !sessionActive {
			mapped := response.MapError(appErr.ErrMissingSession)
			c.AbortWithStatusJSON(mapped.Status, response.ErrorResponse{
				Status: "error",
				Error:  &mapped.Error,
			})
			return
		}

		c.Set(UserPublicIDContextKey, claims.Subject)
		c.Set(SessionIDContextKey, claims.SID)

		c.Next()
	}
}
