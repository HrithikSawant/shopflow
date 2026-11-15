package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port        string
	DatabaseURL string
	RabbitMQURL string
}

// Load reads configuration from the environment. DATABASE_URL and
// RABBITMQ_URL (which carry credentials) have no defaults on purpose: the
// service refuses to start instead of falling back to a built-in password.
func Load() (Config, error) {
	cfg := Config{
		Port:        getenv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RabbitMQURL: os.Getenv("RABBITMQ_URL"),
	}

	var missing []string

	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}

	if cfg.RabbitMQURL == "" {
		missing = append(missing, "RABBITMQ_URL")
	}

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
