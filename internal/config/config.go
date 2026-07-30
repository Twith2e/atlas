package config

import "os"

type Config struct {
	ConnStr            string
	Port               string
	TermiiAPIKey       string
	AccessTokenSecret  string
	RefreshTokenSecret string
	Env                string
}

func Load() *Config {
	return &Config{
		ConnStr:            getEnv("PG_CONN_STRING", ""),
		Port:               getEnv("PORT", "8080"),
		TermiiAPIKey:       getEnv("TERMII_API_KEY", ""),
		AccessTokenSecret:  getEnv("ACCESS_TOKEN_SECRET", ""),
		RefreshTokenSecret: getEnv("REFRESH_TOKEN_SECRET", ""),
		Env:                getEnv("ENV", "development"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return fallback
}
