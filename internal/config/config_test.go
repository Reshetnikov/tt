//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/config --tags=unit -cover -run TestConfig_.*
package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_LoadConfig(t *testing.T) {
	t.Run("TestAllRequiredEnvVarsPresent", func(t *testing.T) {
		os.Setenv("APP_ENV", "production")
		os.Setenv("SITE_URL", "https://example.com")
		os.Setenv("EMAIL_FROM", "noreply@example.com")
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("DB_USER", "user")
		os.Setenv("DB_PASSWORD", "password")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_SSLMODE", "require")
		os.Setenv("REDIS_ADDR", "localhost:6379")
		os.Setenv("MAILGUN_DOMAIN", "example.com")
		os.Setenv("MAILGUN_API_KEY", "api-key")

		cfg := LoadConfig()

		assert.Equal(t, "production", cfg.AppEnv)
		assert.Equal(t, "https://example.com", cfg.SiteUrl)
		assert.Equal(t, "noreply@example.com", cfg.EmailFrom)
		assert.Equal(t, "localhost", cfg.DBHost)
		assert.Equal(t, "5432", cfg.DBPort)
		assert.Equal(t, "user", cfg.DBUser)
		assert.Equal(t, "password", cfg.DBPassword)
		assert.Equal(t, "testdb", cfg.DBName)
		assert.Equal(t, "require", cfg.DBSSLMode)
		assert.Equal(t, "localhost:6379", cfg.RedisAddr)
		assert.Equal(t, "example.com", cfg.MailgunDomain)
		assert.Equal(t, "api-key", cfg.MailgunApiKey)
	})

	t.Run("TestMissingRequiredEnvVars", func(t *testing.T) {
		os.Setenv("SITE_URL", "")
		os.Setenv("DB_HOST", "")
		os.Setenv("DB_PORT", "")
		os.Setenv("DB_USER", "")
		os.Setenv("DB_PASSWORD", "")
		os.Setenv("DB_NAME", "")
		os.Setenv("DB_SSLMODE", "")
		os.Setenv("REDIS_ADDR", "")
		os.Setenv("MAILGUN_DOMAIN", "")
		os.Setenv("MAILGUN_API_KEY", "")

		cfg := LoadConfig()

		assert.Empty(t, cfg.SiteUrl)
		assert.Empty(t, cfg.DBHost)
		assert.Empty(t, cfg.DBPort)
		assert.Empty(t, cfg.DBUser)
		assert.Empty(t, cfg.DBPassword)
		assert.Empty(t, cfg.DBName)
		assert.Empty(t, cfg.DBSSLMode)
		assert.Empty(t, cfg.RedisAddr)
		assert.Empty(t, cfg.MailgunDomain)
		assert.Empty(t, cfg.MailgunApiKey)
	})
}

func TestConfig_GetPostgresDSN(t *testing.T) {
	t.Run("TestGetPostgresDSN", func(t *testing.T) {
		os.Setenv("DB_USER", "user")
		os.Setenv("DB_PASSWORD", "password")
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_SSLMODE", "require")

		cfg := LoadConfig()

		expectedDSN := "postgres://user:password@localhost:5432/testdb?sslmode=require"
		assert.Equal(t, expectedDSN, cfg.GetPostgresDSN())
	})

	t.Run("TestGetPostgresDSNWithEmptySSLMode", func(t *testing.T) {
		os.Setenv("DB_USER", "user")
		os.Setenv("DB_PASSWORD", "password")
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_SSLMODE", "")

		cfg := LoadConfig()

		expectedDSN := "postgres://user:password@localhost:5432/testdb?sslmode="
		assert.Equal(t, expectedDSN, cfg.GetPostgresDSN())
	})
}
