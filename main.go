package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/xedom/codeduel-lobby/codeduel"
	"github.com/xedom/codeduel-lobby/codeduel/config"
	"github.com/xedom/codeduel-lobby/codeduel/utils"
)

var lobbies = make(map[string]*codeduel.Lobby)

// var addr = flag.String("addr", ":8080", "http service address")

func main() {
	// loading env only if not in production
	if utils.GetEnv("ENV", "development") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("[MAIN] Error loading .env file")
		}
	}
	config := config.LoadConfig()
	runner := codeduel.NewRunner(config.RunnerURL)
	backend := codeduel.NewBackend(config.BackendURL, config.BackendApiKey)
	server := codeduel.NewApiServer(config, lobbies, &runner, &backend)
	server.Run()
}
