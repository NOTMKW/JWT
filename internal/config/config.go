package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret    string
	Port         string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
}

func Load() *Config {
	godotenv.Load("cmd/.env")

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	return &Config{
		JWTSecret: getEnv("JWT_SECRET", "default_secret"),
		Port:      getEnv("PORT", "8080"),
		SMTPHost:	getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:	smtpPort,
		SMTPUsername: getEnv("SMTP_USERNAME",""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail: getEnv("FROM_EMAIL", ""),
	}
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return value
	}
	return defaultValue
}
