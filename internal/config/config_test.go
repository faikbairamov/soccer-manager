package config_test

import (
	"os"
	"testing"

	"github.com/faikbairamov/soccer-manager/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var requiredEnvVars = []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME", "JWT_SECRET"}

func setRequired(t *testing.T) {
	t.Helper()
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_USER", "soccer")
	t.Setenv("DB_PASSWORD", "soccer")
	t.Setenv("DB_NAME", "soccer_manager")
	t.Setenv("JWT_SECRET", "secret")
}

func TestLoad_Defaults(t *testing.T) {
	setRequired(t)

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.AppPort)
	assert.Equal(t, "development", cfg.AppEnv)
	assert.Equal(t, 5432, cfg.DBPort)
	assert.Equal(t, "disable", cfg.DBSSLMode)
	assert.Equal(t, int32(25), cfg.DBMaxConns)
	assert.Equal(t, int32(5), cfg.DBMinConns)
	assert.Equal(t, 72, cfg.JWTExpiryHours)
}

func TestLoad_CustomPort(t *testing.T) {
	setRequired(t)
	t.Setenv("APP_PORT", "9090")

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "9090", cfg.AppPort)
}

func TestLoad_MissingRequired(t *testing.T) {
	for _, key := range requiredEnvVars {
		if original, exists := os.LookupEnv(key); exists {
			t.Cleanup(func() { os.Setenv(key, original) })
		} else {
			t.Cleanup(func() { os.Unsetenv(key) })
		}
		os.Unsetenv(key)
	}
	_, err := config.Load()
	require.Error(t, err)
}
