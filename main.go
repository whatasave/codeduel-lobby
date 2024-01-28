package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/xedom/codeduel-lobby/types"
)

var lobbies = make(map[string]*Lobby)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()
	fmt.Printf("Starting server on addr http://%s\n", *addr)

	router := mux.NewRouter()
	router.HandleFunc("/ws/{lobby}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("\n--- New connection ---\n")

		lobbyTab := mux.Vars(r)["lobby"]
		lobby, ok := lobbies[lobbyTab]
		if !ok {
			lobby = newLobby()
			lobbies[lobbyTab] = lobby
			go lobby.run()
		}

		// fmt.Printf("Current lobbies: %v\n", lobbies)
		if len(lobby.clients) >= lobby.maxPlayers {
			fmt.Printf("Lobby %s is full\n", lobbyTab)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if lobby.status != types.STARTING {
			fmt.Printf("Lobby %s is not in starting state\n", lobbyTab)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if lobby.isLocked { // TODO: add password check
			fmt.Printf("Lobby %s is locked\n", lobbyTab)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if lobby.endTimestamp != 0 && lobby.endTimestamp < int(time.Now().Unix()) {
			fmt.Printf("Time is up for lobby %s\n", lobbyTab)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		err := serveWs(lobby, w, r, lobbyTab)
		if err != nil {
			log.Println(err)
		}

		for currentLobbyTag, currentLobbyRef := range lobbies {
			fmt.Printf("Lobby %s:\n", currentLobbyTag)
			for client, ok := range currentLobbyRef.clients {
				fmt.Printf("Client %s: %s (owner %v) - %v\n", client.ID, client.Token, client.Owner, ok)
			}
		}

	})
	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
	}

	frontendUrl := "http://localhost:5173"

	err := http.ListenAndServe(server.Addr, handlers.CORS(
		handlers.AllowedOrigins([]string{frontendUrl}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "PUT", "DELETE"}),
		handlers.AllowedHeaders([]string{
			"Content-Type",
			"Access-Control-Allow-Headers",
			"Authorization",
			"X-Requested-With",
			"x-jwt-token",
		}),
		handlers.AllowCredentials(),
	)(router))

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
