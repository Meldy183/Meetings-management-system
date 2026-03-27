package config

import (
	"errors"
	"os"
)

type Config struct {
	HTTPAddr      string
	DBDSN         string
	Env           string
	AdminPassword string
	APIKey        string
}

func Load() (*Config, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://localhost/meetings_editor?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		return nil, errors.New("ADMIN_PASSWORD env var is required")
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, errors.New("API_KEY env var is required")
	}

	return &Config{
		HTTPAddr:      ":" + port,
		DBDSN:         dsn,
		Env:           env,
		AdminPassword: adminPassword,
		APIKey:        apiKey,
	}, nil
}
