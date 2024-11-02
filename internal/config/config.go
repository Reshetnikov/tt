package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	AppEnv      string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	RedisAddr   string
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() (*Config) {
	return &Config{
		Port:       os.Getenv("PORT"),
		AppEnv:     os.Getenv("APP_ENV"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		RedisAddr:  os.Getenv("REDIS_ADDR"),
	}
}

func (cfg *Config) GetPostgresDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", 
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
}