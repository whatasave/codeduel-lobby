package utils

import (
	"log"
	"os"
)

type Config struct {
	Host string
	Port string

	CorsOrigin      string
	CorsMethods     string
	CorsHeaders     string
	CorsCredentials bool

	BackendURL    string
	BackendApiKey string

	RunnerURL    string
	RunnerApiKey string
}

func LoadConfig() *Config {
	return &Config{
		Host: GetEnv("HOST", "localhost"),
		Port: GetEnv("PORT", "5010"),

		CorsOrigin:      GetEnv("CORS_ORIGIN", "http://localhost:5173"),
		CorsMethods:     GetEnv("CORS_METHODS", "POST"),
		CorsHeaders:     GetEnv("CORS_HEADERS", "Content-Type, x-token, Accept, Content-Length, Accept-Encoding, Authorization,X-CSRF-Token"),
		CorsCredentials: GetEnv("CORS_CREDENTIALS", "true") == "true",

		BackendURL:    GetEnv("BACKEND_URL", "http://localhost:5000"),
		BackendApiKey: GetEnv("BACKEND_API_KEY", "xxx"),

		RunnerURL:    GetEnv("RUNNER_URL", "http://localhost:5020"),
		RunnerApiKey: GetEnv("RUNNER_API_KEY", "xxx"),
	}
}

func GetEnv(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Printf("[WARN] Environment variable %s not found, using default value %s\n", key, defaultValue)
		return defaultValue
	}

	return value
}
