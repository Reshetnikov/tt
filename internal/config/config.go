package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppEnv     string
	SiteUrl    string
	EmailFrom  string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	RedisAddr  string
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() *Config {
	return &Config{
		AppEnv:     os.Getenv("APP_ENV"),
		SiteUrl:    os.Getenv("SITE_URL"),
		EmailFrom:  os.Getenv("EMAIL_FROM"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSSLMode:  os.Getenv("DB_SSLMODE"),
		RedisAddr:  os.Getenv("REDIS_ADDR"),
	}
}

func (cfg *Config) GetPostgresDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)
}
