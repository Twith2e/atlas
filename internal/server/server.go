package server

import (
	"atlas/internal/config"
	"net/http"
	"time"
)

func NewServer(cfg *config.Config) (*http.Server, error) {
	router, err := NewRouter(cfg)
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}, nil
}
