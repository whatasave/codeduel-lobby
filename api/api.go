package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/xedom/codeduel-lobby/codeduel"
)

type APIServer struct {
	Addr              string
	Lobbies           map[string]*codeduel.Lobby
	ReadHeaderTimeout time.Duration
}

type VerifyTokenResponse struct {
	Id			int32  `json:"id"`
	Username	string `json:"username"`
	Email		string `json:"email"`
	Image_url	string `json:"image_url"`
	Expires_at	string `json:"expires_at"`
}

func NewAPIServer(addr string, lobbies map[string]*codeduel.Lobby) *APIServer {
	log.Print("[API] Starting API server on ", addr)
	return &APIServer{
		Addr:              addr,
		Lobbies:           lobbies,
		ReadHeaderTimeout: 3 * time.Second,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/create", s.createLobby)
	router.HandleFunc("/lobbies", s.getAllLobbies)
	router.HandleFunc("/join/{lobby}", s.joinLobby)
	router.HandleFunc("/connect/{lobby}", s.connectLobby)

	err := http.ListenAndServe(s.Addr, handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"POST"}),
		handlers.AllowedHeaders([]string{}),
		handlers.AllowCredentials(),
	)(router))

	if err != nil {
		log.Fatal("[API] Cannot start http server: ", err)
	}
}

func (s *APIServer) createLobby(response http.ResponseWriter, request *http.Request) {
	user, err := GetUser(request)
	fmt.Println("user: ", user)
	if err != nil {
		codeduel.RejectConnection(response, request, codeduel.Unauthorized, err.Error())
		return
	}
	lobby := codeduel.NewLobby(user)
	s.Lobbies[lobby.Id] = &lobby
	_, err = codeduel.StartWebSocket(response, request, &lobby, user)
	if err != nil {
		codeduel.RejectConnection(response, request, codeduel.InternalServerError, "cannot start websocket connection")
		return
	}
}

func (s *APIServer) getAllLobbies(response http.ResponseWriter, request *http.Request) {
	type lobbyListType struct {
		Id          string   							`json:"id"`
		Owner       *codeduel.User 						`json:"owner"`
		Users       map[codeduel.UserId]*codeduel.User 	`json:"users"`
		Max_players int      							`json:"max_players"`
		State       any      							`json:"state"`
	}

	lobbyList := make([]lobbyListType, 0, len(s.Lobbies))

	for key, lobby := range s.Lobbies {

		lobbyUsers := make([]string, 0, len(lobby.Users))

		for userID := range lobby.Users {
			lobbyUsers = append(lobbyUsers, strconv.Itoa(int(userID)))
		}

		lobbyList = append(lobbyList, lobbyListType{
			Id:          key,
			// Owner:       strconv.Itoa(int(lobby.Owner.Id)), // TODO replace with the name of the owner of the lobby
			Owner:       lobby.Owner,
			Users:       lobby.Users,
			Max_players: lobby.Settings.MaxPlayers,
			State:       lobby.State,
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

func (s *APIServer) joinLobby(response http.ResponseWriter, request *http.Request) {
	user, err := GetUser(request)
	if err != nil {
		codeduel.RejectConnection(response, request, codeduel.Unauthorized, err.Error())
		return
	}
	lobbyId := mux.Vars(request)["lobby"]
	lobby, ok := s.Lobbies[lobbyId]
	if !ok {
		response.WriteHeader(http.StatusNotFound)
		return
	}
	if err := lobby.CannotJoin(user); err != nil {
		response.WriteHeader(http.StatusForbidden)
		return
	}
	if isUserInLobby := lobby.GetUser(user); isUserInLobby == nil {
		lobby.AddUser(user)
	}
	_, err = codeduel.StartWebSocket(response, request, lobby, user)
	if err != nil {
		codeduel.RejectConnection(response, request, codeduel.InternalServerError, "cannot start websocket connection")
		return
	}
}

func (s *APIServer) connectLobby(response http.ResponseWriter, request *http.Request) {
	user, err := GetUser(request)
	if err != nil {
		codeduel.RejectConnection(response, request, codeduel.Unauthorized, err.Error())
		return
	}
	lobbyId := mux.Vars(request)["lobby"]
	lobby, ok := s.Lobbies[lobbyId]
	if !ok {
		codeduel.RejectConnection(response, request, codeduel.NotFound, "lobby not found")
		return
	}
	if user := lobby.GetUser(user); user == nil {
		codeduel.RejectConnection(response, request, codeduel.Forbidden, "user not in lobby")
		return
	}
	_, err = codeduel.StartWebSocket(response, request, lobby, user)
	if err != nil {
		codeduel.RejectConnection(response, request, codeduel.InternalServerError, "cannot start websocket connection")
		return
	}
}

func GetUser(request *http.Request) (*codeduel.User, error) {
	cookie, err := request.Cookie("jwt")
	if err != nil {
		return nil, errors.New("missing jwt cookie")
	}
	// id, err := strconv.Atoi(cookie.Value)
	// if err != nil { return nil, err }

	// TODO: validate jwt calling codeduel-be
	verifyTokenResponse, err := verifyJwt(cookie.Value)


	return &codeduel.User{
		Id: codeduel.UserId(verifyTokenResponse.Id),
		Username: verifyTokenResponse.Username,
		Email: verifyTokenResponse.Email,
		Avatar: verifyTokenResponse.Image_url,
		Token: cookie.Value,
		TokenExpiresAt: verifyTokenResponse.Expires_at,
	}, nil
}

func verifyJwt(jwt string) (*VerifyTokenResponse, error) {
	requestURL := "http://localhost:5000/api/v1/validateToken"
	requestBodyMap := map[string]string{
		"token": jwt,
	}
	requestBody, err := json.Marshal(requestBodyMap)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Invalid token")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	verifyTokenResponse := &VerifyTokenResponse{}
	json.Unmarshal(respBody, &verifyTokenResponse)

	return verifyTokenResponse, nil
}