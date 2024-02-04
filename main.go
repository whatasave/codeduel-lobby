package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/xedom/codeduel-lobby/codeduel"
)

var lobbies = make(map[string]*codeduel.Lobby)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()
	log.Printf("starting server on addr http://%s\n", *addr)

	router := mux.NewRouter()
	router.HandleFunc("/create", createLobby)
	router.HandleFunc("/lobbies", getAllLobbies)
	router.HandleFunc("/join/{lobby}", joinLobby)
	router.HandleFunc("/connect/{lobby}", connectLobby)

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
	user, err := GetUser(request)
	if err != nil {
		codeduel.RejectConnection(response, request, codeduel.Unauthorized, err.Error())
		return
	}
	lobby := codeduel.NewLobby(user)
	lobbies[lobby.Id] = &lobby
	_, err = codeduel.StartWebSocket(response, request, &lobby, user)
	if err != nil {
		codeduel.RejectConnection(response, request, codeduel.InternalServerError, "cannot start websocket connection")
		return
	}
}

func getAllLobbies(response http.ResponseWriter, request *http.Request) {
	type lobbyListType struct {
		Id          string   `json:"id"`
		Owner       string   `json:"owner"`
		Users       []string `json:"players"`
		Max_players int      `json:"max_players"`
		State       string   `json:"state"`
	}

	lobbyList := make([]lobbyListType, 0, len(lobbies))

	for key, lobby := range lobbies {

		lobbyUsers := make([]string, 0, len(lobby.Users))

		for userID := range lobby.Users {
			lobbyUsers = append(lobbyUsers, strconv.Itoa(int(userID)))
		}

		lobbyList = append(lobbyList, lobbyListType{
			Id:          key,
			Owner:       strconv.Itoa(int(lobby.Owner.Id)), // TODO replace with the name of the owner of the lobby
			Users:       lobbyUsers,
			Max_players: lobby.Settings.MaxPlayers,
			State:       "PreLobby", // TODO get lobby status
		})
	}

	fmt.Printf("lobbies: %v\n", lobbyList)

	response.Header().Add("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err := json.NewEncoder(response).Encode(lobbyList)
	if err != nil {
		log.Fatalf("[API] error with encoding lobbies into json: %v", err.Error())
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func joinLobby(response http.ResponseWriter, request *http.Request) {
	user, err := GetUser(request)
	if err != nil {
		codeduel.RejectConnection(response, request, codeduel.Unauthorized, err.Error())
		return
	}
	lobbyId := mux.Vars(request)["lobby"]
	lobby, ok := lobbies[lobbyId]
	if !ok {
		response.WriteHeader(http.StatusNotFound)
		return
	}
	if err := lobby.CannotJoin(user); err != nil {
		response.WriteHeader(http.StatusForbidden)
		return
	}
	lobby.AddUser(user)
	_, err = codeduel.StartWebSocket(response, request, lobby, user)
	if err != nil {
		codeduel.RejectConnection(response, request, codeduel.InternalServerError, "cannot start websocket connection")
		return
	}
}

func connectLobby(response http.ResponseWriter, request *http.Request) {
	user, err := GetUser(request)
	if err != nil {
		codeduel.RejectConnection(response, request, codeduel.Unauthorized, err.Error())
		return
	}
	lobbyId := mux.Vars(request)["lobby"]
	lobby, ok := lobbies[lobbyId]
	if !ok {
		codeduel.RejectConnection(response, request, codeduel.NotFound, "lobby not found")
		return
	}
	if user := lobby.GetUser(user); user == nil {
		codeduel.RejectConnection(response, request, codeduel.Forbidden, "user not in lobby")
		return
	}
	connection, err := codeduel.StartWebSocket(response, request, lobby, user)
	if err != nil {
		codeduel.RejectConnection(response, request, codeduel.InternalServerError, "cannot start websocket connection")
		return
	} else {
		user.Connection = connection
	}
}

func GetUser(request *http.Request) (*codeduel.User, error) {
	cookie, err := request.Cookie("jwt")
	if err != nil {
		return nil, errors.New("missing jwt cookie")
	}
	// TODO: validate jwt calling codeduel-be
	id, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return nil, err
	}
	return &codeduel.User{
		Id: codeduel.UserId(id),
	}, nil
}
