package config

import "github.com/xedom/codeduel-lobby/codeduel/utils"

type Config struct {
	Host string
	Port string

	BackendURL    string
	BackendApiKey string

	RunnerURL    string
	RunnerApiKey string
}

func LoadConfig() *Config {
	return &Config{
		Host: utils.GetEnv("HOST", "localhost"),
		Port: utils.GetEnv("PORT", "5010"),

		BackendURL:    utils.GetEnv("BACKEND_URL", "http://localhost:5000"),
		BackendApiKey: utils.GetEnv("BACKEND_API_KEY", "xxx"),

		RunnerURL:    utils.GetEnv("RUNNER_URL", "http://localhost:5020"),
		RunnerApiKey: utils.GetEnv("RUNNER_API_KEY", "xxx"),
	}
}
