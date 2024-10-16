package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
}

// Boot order:
// 1: .env.development.local
// 2: .env.local
// 3: .env.development
// 4: .env
//
// 1: .env.test.local
// 2: .env.test
// 3: .env
func LoadConfig() (*Config, error) {
	env := os.Getenv("TIME_TRACKER_ENV")
	if env == "" {
		env = "development"
	}

	godotenv.Load(".env." + env + ".local")
	if env != "test" {
		godotenv.Load(".env.local")
	}
	godotenv.Load(".env." + env)
	godotenv.Load()

	cfg := &Config{
		Port: getEnv("PORT", "8080"),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
