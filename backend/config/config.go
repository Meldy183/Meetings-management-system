package config

import (
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
		adminPassword = "admin"
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		apiKey = "admin"
	}

	return &Config{
		HTTPAddr:      ":" + port,
		DBDSN:         dsn,
		Env:           env,
		AdminPassword: adminPassword,
		APIKey:        apiKey,
	}, nil
}
