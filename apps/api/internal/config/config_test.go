package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/HrithikSawant/shopflow-api/internal/config"
)

func TestLoadRequiresCredentialsFromEnvironment(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("RABBITMQ_URL", "")

	_, err := config.Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL")
	assert.Contains(t, err.Error(), "RABBITMQ_URL")
}

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://db")
	t.Setenv("RABBITMQ_URL", "amqp://mq")
	t.Setenv("PORT", "")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "postgres://db", cfg.DatabaseURL)
	assert.Equal(t, "amqp://mq", cfg.RabbitMQURL)
	assert.Equal(t, "8080", cfg.Port, "port keeps its default")
}
