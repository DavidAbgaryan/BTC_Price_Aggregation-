package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	PollInterval   time.Duration
	RequestTimeout time.Duration
	MaxRetries     int
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		PollInterval:   getEnvDuration("POLL_INTERVAL", 10*time.Second),
		RequestTimeout: getEnvDuration("REQUEST_TIMEOUT", 3*time.Second),
		MaxRetries:     getEnvInt("MAX_RETRIES", 3),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if val, exists := os.LookupEnv(key); exists {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
