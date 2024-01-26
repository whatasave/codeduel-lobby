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

// type Client struct {
// 	ID   string `json:"id"`
// 	// Conn *Conn

// 	Name string `json:"name"`
// 	Lobby *Lobby
// }

// type Lobby struct {
//   ID      string
//   Clients []*Client
// }

var lobbies = make(map[string]*Lobby)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()
	fmt.Printf("Starting server on addr http://%s\n", *addr)

	router := mux.NewRouter()
	router.HandleFunc("/ws/{lobby}", func(w http.ResponseWriter, r *http.Request) {
		lobbyTab := mux.Vars(r)["lobby"]
		lobby, ok := lobbies[lobbyTab]
		if !ok {
			lobby = newLobby()
			lobbies[lobbyTab] = lobby
			go lobby.run()
		}

		fmt.Printf("Current lobbies: %v\n", lobbies)

		err := serveWs(lobby, w, r, lobbyTab)
		if err != nil { log.Println(err) }
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

	if err != nil { log.Fatal("ListenAndServe: ", err) }
}