package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	AppName     string
	Environment string
	Port        string
}

func LoadConfig() (*Config, error) {
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
