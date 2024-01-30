package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/xedom/codeduel-lobby/codeduel"
)

var lobbies = make(map[string]*codeduel.Lobby)
var connections = make(map[*codeduel.UserId]*websocket.Conn)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()
	log.Printf("starting server on addr http://%s\n", *addr)

	router := mux.NewRouter()
	router.HandleFunc("/create", createLobby)
	router.HandleFunc("/join/{lobby}", joinLobby)

	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
	}

	err := http.ListenAndServe(server.Addr, handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"POST"}),
		handlers.AllowedHeaders([]string{}),
		handlers.AllowCredentials(),
	)(router))

	if err != nil {
		log.Fatal("cannot start http server: ", err)
	}
}

func createLobby(response http.ResponseWriter, request *http.Request) {
	user := GetUser(request)
	if user == nil {
		response.WriteHeader(http.StatusUnauthorized)
		return
	}
	lobbyId := uuid.NewString()
	lobby := codeduel.NewLobby(user)
	lobbies[lobbyId] = &lobby
	connection, err := codeduel.StartWebSocket(response, request, &lobby, user)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	connections[&user.Id] = connection
}

func joinLobby(response http.ResponseWriter, request *http.Request) {
	user := GetUser(request)
	if user == nil {
		response.WriteHeader(http.StatusUnauthorized)
		return
	}
	lobbyId := mux.Vars(request)["lobby"]
	lobby, ok := lobbies[lobbyId]
	if !ok {
		response.WriteHeader(http.StatusNotFound)
		return
	}
	if !lobby.CanJoin(user) {
		response.WriteHeader(http.StatusForbidden)
		return
	}
	lobby.AddUser(user)
	connection, err := codeduel.StartWebSocket(response, request, lobby, user)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	connections[&user.Id] = connection
}

func GetUser(request *http.Request) *codeduel.User {
	cookie, err := request.Cookie("jwt")
	if err != nil {
		return nil
	}
	// TODO: validate jwt calling codeduel-be
	id, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return nil
	}
	return &codeduel.User{
		Id: codeduel.UserId(id),
	}
}
