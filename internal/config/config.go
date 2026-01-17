package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
	AIAPIKey    string
}

func LoadConfig() *Config {
	return &Config{
		DatabaseURL: mustEnv("DATABASE_URL"),
		Port:        envOr("PORT", "8080"),
		AIAPIKey:    os.Getenv("AI_API_KEY"),
	}
}

func mustEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("missing env:" + key)
	}
	return value
}

func envOr(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
