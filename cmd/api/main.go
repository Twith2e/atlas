package main

import (
	"atlas/internal/config"
	"atlas/internal/server"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
)

func run(ctx context.Context) error {
	errChan := make(chan error, 1)
	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	srv, err := server.NewServer(cfg)
	if err != nil {
		return err
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer shutdownCancel()
	err = srv.Shutdown(shutdownCtx)
	defer sentry.Flush(2 * time.Second)
	return err
}

// @title           Atlas API
// @version         1.0
// @description     API documentation for Atlas.
// @BasePath        /api/v1
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found (using system environment)")
	}

	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
