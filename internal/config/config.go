package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret string
	Port      string
}

func Load() (*Config) {
	godotenv.Load("cmd/.env")

	return &Config{
		JWTSecret: getEnv("JWT_SECRET", "default_secret"),
		Port:      getEnv("PORT", "8080"),
	}
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return value
	}
	return defaultValue
}
