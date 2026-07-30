package auth

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, authGuard, sessionGuard, activeSessionGuard gin.HandlerFunc, authHandler *Handler) {
	authGroup := r.Group("/auth")

	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/logout", sessionGuard, authHandler.Logout)
		authGroup.POST("/refresh", sessionGuard, activeSessionGuard, authHandler.RefreshAccessToken)
	}

}
