package server

import (
	"atlas/internal/config"
	"atlas/internal/database"
	"atlas/internal/modules/auth"
	"atlas/internal/providers/tokens"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "atlas/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(cfg *config.Config) (*gin.Engine, error) {
	connStr := cfg.ConnStr
	if connStr == "" {
		return nil, fmt.Errorf("db connection string not set")
	}

	db, err := connectWithPostgres(connStr, 5, time.Second)
	if err != nil {
		return nil, err
	}

	r := gin.Default()

	r.HEAD("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.StaticFile("/favicon.ico", "./static/gopher.jpg")

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	apiV1 := api.Group("/v1")

	tokenGenerator := tokens.NewJWT(cfg.AccessTokenSecret, cfg.RefreshTokenSecret)

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, db, tokenGenerator)
	authHandler := auth.NewHandler(authService)
	auth.RegisterRoutes(apiV1, authHandler)

	return r, nil
}

func connectWithPostgres(connStr string, attempts int, baseDelay time.Duration) (*sql.DB, error) {
	if attempts <= 0 {
		attempts = 1
	}

	if baseDelay <= 0 {
		baseDelay = time.Second
	}

	var lastErr error

	for attempt := 1; attempt <= attempts; attempt++ {
		db, err := database.NewPostgresDB(connStr)
		if err == nil {
			return db, nil
		}

		lastErr = err
		if attempt == attempts {
			break
		}

		delay := baseDelay * time.Duration(attempt)
		log.Printf("database connection attempt %d/%d failed: %v (retrying in %s)", attempt, attempts, err, delay)
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("database connection failed after %d attempts: %w", attempts, lastErr)
}
