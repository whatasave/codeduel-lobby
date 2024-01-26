package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
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

		err := serveWs(lobby, w, r, lobbyTab)
		if err != nil {
			log.Println(err)
		}

		for currLobbyTag, currentLobbyRef := range lobbies {
			fmt.Printf("Lobby %s:\n", currLobbyTag)
			for client, ok := range currentLobbyRef.clients {
				fmt.Printf("Clien %s: %s - %v\n", client.ID, client.Token, ok)
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
