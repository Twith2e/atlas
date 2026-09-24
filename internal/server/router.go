package server

import (
	"atlas/internal/config"
	"atlas/internal/database"
	"atlas/internal/middleware"
	"atlas/internal/modules/auth"
	"atlas/internal/providers/tokens"
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	_ "atlas/docs"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	sentryslog "github.com/getsentry/sentry-go/slog"
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

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Debug:            true,
		DataCollection:   &sentry.DataCollection{},
		EnableTracing:    true,
		TracesSampleRate: 0.1,
		BeforeSendLog: func(log *sentry.Log) *sentry.Log {
			if log.Level == sentry.LogLevelTrace {
				return nil
			}

			return log
		},
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
		return nil, err
	}

	ctx := context.Background()
	handler := sentryslog.Option{
		LogLevel:  []slog.Level{slog.LevelInfo, slog.LevelWarn, slog.LevelError},
		AddSource: true,
	}.NewSentryHandler(ctx)

	logger := slog.New(handler)
	slog.SetDefault(logger)

	app := gin.Default()

	app.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))

	app.HEAD("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	app.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	app.StaticFile("/favicon.ico", "./static/gopher.jpg")

	app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := app.Group("/api")
	apiV1 := api.Group("/v1")

	tokenGenerator := tokens.NewJWT(cfg.AccessTokenSecret)

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, db, tokenGenerator, tokens.GenerateRefreshToken())
	authHandler := auth.NewHandler(authService, cfg.Env)
	auth.RegisterRoutes(apiV1, middleware.AuthMiddleware(tokenGenerator, authService), middleware.SessionGuard(tokenGenerator, authService), middleware.RequireActiveSession(authService), authHandler)

	return app, nil
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
