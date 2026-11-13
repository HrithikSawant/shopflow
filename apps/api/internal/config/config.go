package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	RabbitMQURL string
}

func Load() Config {
	return Config{
		Port:        getenv("PORT", "8080"),
		DatabaseURL: getenv("DATABASE_URL", ""),
		RabbitMQURL: getenv("RABBITMQ_URL", ""),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
