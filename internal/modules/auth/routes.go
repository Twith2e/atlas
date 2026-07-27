package auth

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, authHandler *Handler) {
	authGroup := r.Group("/auth")

	{
		authGroup.POST("/register", authHandler.Register)
		// authGroup.POST("/login", authHandler.Login)
		// authGroup.POST("/logout", authHandler.Logout)
		// authGroup.POST("/refresh", authHandler.Refresh)
	}

}
