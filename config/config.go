package config

// implement me
type Config struct {
	HTTPAddr string
	DBDSN    string
}

func Load() (*Config, error) {
	// implement me
	return &Config{
		HTTPAddr: ":8080",
		DBDSN:    "postgres://localhost/meetings_editor?sslmode=disable",
	}, nil
}
