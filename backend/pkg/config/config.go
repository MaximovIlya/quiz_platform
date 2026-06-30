package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string

	JWTSecret          string
	JWTAccessTokenTTL  string // e.g. "15m"
	JWTRefreshTokenTTL string // e.g. "168h"

	GoogleClientID     string
	GoogleClientSecret string

	ResendAPIKey string

	VercelBlobToken string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, reading from environment")
	}

	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: mustGetEnv("DATABASE_URL"),

		JWTSecret:          mustGetEnv("JWT_SECRET"),
		JWTAccessTokenTTL:  getEnv("JWT_ACCESS_TTL", "15m"),
		JWTRefreshTokenTTL: getEnv("JWT_REFRESH_TTL", "168h"),

		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),

		ResendAPIKey: os.Getenv("RESEND_API_KEY"),

		VercelBlobToken: os.Getenv("BLOB_READ_WRITE_TOKEN"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env variable %s is not set", key)
	}
	return v
}
