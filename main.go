package main

import (
	"flag"
	"log"

	"github.com/xedom/codeduel-lobby/api"
	"github.com/xedom/codeduel-lobby/codeduel"
)

var lobbies = make(map[string]*codeduel.Lobby)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()
	log.Printf("starting server on addr http://%s\n", *addr)

	server := api.NewAPIServer(*addr, lobbies)
	server.Run()
}
