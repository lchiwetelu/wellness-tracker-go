package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env             string
	HTTPPort        string
	DatabaseURL     string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	FrontendUrl     string
}

// Load reads configuration from environment variables, optionally loading them
// from a .env file in development.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("config: no .env file found (this is fine in production): %v", err)
	}

	return &Config{
		Env:             getEnv("APP_ENV", "development"),
		HTTPPort:        getEnv("PORT", "8080"),
		DatabaseURL:     mustGetEnv("DATABASE_URL"),
		ReadTimeout:     getEnvDuration("HTTP_READ_TIMEOUT", 5*time.Second),
		WriteTimeout:    getEnvDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
		ShutdownTimeout: getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		FrontendUrl:     getEnv("FRONTEND_URL", "frontend"),
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
		log.Fatalf("config: required environment variable %q is not set", key)
	}
	return v
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			log.Printf("config: invalid duration for %s=%q, using fallback %s: %v", key, v, fallback, err)
			return fallback
		}
		return d
	}
	return fallback
}
