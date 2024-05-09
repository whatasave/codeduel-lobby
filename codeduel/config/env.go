package config

import "github.com/xedom/codeduel-lobby/codeduel/utils"

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
		Host: utils.GetEnv("HOST", "localhost"),
		Port: utils.GetEnv("PORT", "5010"),

		CorsOrigin:      utils.GetEnv("CORS_ORIGIN", "http://localhost:5173"),
		CorsMethods:     utils.GetEnv("CORS_METHODS", "POST"),
		CorsHeaders:     utils.GetEnv("CORS_HEADERS", "Content-Type, x-token, Accept, Content-Length, Accept-Encoding, Authorization,X-CSRF-Token"),
		CorsCredentials: utils.GetEnv("CORS_CREDENTIALS", "true") == "true",

		BackendURL:    utils.GetEnv("BACKEND_URL", "http://localhost:5000"),
		BackendApiKey: utils.GetEnv("BACKEND_API_KEY", "xxx"),

		RunnerURL:    utils.GetEnv("RUNNER_URL", "http://localhost:5020"),
		RunnerApiKey: utils.GetEnv("RUNNER_API_KEY", "xxx"),
	}
}
