package main

import (
	"github.com/xedom/codeduel-lobby/codeduel"
	"github.com/xedom/codeduel-lobby/codeduel/utils"
)

var lobbies = make(map[string]*codeduel.Lobby)

// var addr = flag.String("addr", ":8080", "http service address")

func main() {
	config := utils.LoadConfig()
	runner := codeduel.NewRunner(config.RunnerURL)
	backend := codeduel.NewBackend(config.BackendURL, config.BackendApiKey)
	server := codeduel.NewApiServer(config, lobbies, &runner, &backend)
	server.Run()
}
