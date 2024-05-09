package codeduel

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/xedom/codeduel-lobby/codeduel/config"
	"github.com/xedom/codeduel-lobby/codeduel/utils"
)

type APIServer struct {
	Config            *config.Config
	Addr              string
	Lobbies           map[string]*Lobby
	ReadHeaderTimeout time.Duration
	Runner            *Runner
	Backend           *Backend
}

type VerifyTokenResponse struct {
	Id        int32  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"expires_at"`
}

func NewApiServer(config *config.Config, lobbies map[string]*Lobby, runner *Runner, backend *Backend) *APIServer {
	address := fmt.Sprintf("%s:%s", config.Host, config.Port)
	log.Print("[API] Starting API server on http://", address)
	return &APIServer{
		Config:            config,
		Addr:              address,
		Lobbies:           lobbies,
		ReadHeaderTimeout: 3 * time.Second,
		Runner:            runner,
		Backend:           backend,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/health", s.healthCheck)
	router.HandleFunc("/create", s.createLobby)
	router.HandleFunc("/lobbies", s.getAllLobbies)
	router.HandleFunc("/join/{lobby}", s.joinLobby)
	router.HandleFunc("/connect/{lobby}", s.connectLobby)

	err := http.ListenAndServe(s.Addr, handlers.CORS(
		handlers.AllowedOrigins([]string{s.Config.CorsOrigin}),
		handlers.AllowedMethods([]string{s.Config.CorsMethods}),
		handlers.AllowedHeaders([]string{s.Config.CorsHeaders}),
		handlers.AllowCredentials(),
	)(router))

	if err != nil {
		log.Fatal("[API] Cannot start http server: ", err)
	}
}

func (s *APIServer) healthCheck(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(map[string]string{"status": "ok"})
}

func (s *APIServer) createLobby(response http.ResponseWriter, request *http.Request) {
	user, err := s.GetUser(request)
	fmt.Println("user: ", user)
	if err != nil {
		log.Printf("[API] error getting user: %v", err)
		_ = RejectConnection(response, request, Unauthorized, err.Error())
		return
	}
	languages, err := s.Runner.AvailableLanguages()
	if err != nil {
		log.Printf("[API] error getting available languages: %v", err)
		_ = RejectConnection(response, request, InternalServerError, "cannot contact runner")
		return
	}
	lobby := NewLobby(user, languages)
	s.Lobbies[lobby.Id] = &lobby
	_, err = s.StartWebSocket(response, request, &lobby, user)
	if err != nil {
		log.Printf("[API] error starting websocket: %v", err)
		_ = RejectConnection(response, request, InternalServerError, "cannot start websocket connection")
		return
	}
}

func (s *APIServer) joinLobby(response http.ResponseWriter, request *http.Request) {
	user, err := s.GetUser(request)
	if err != nil {
		_ = RejectConnection(response, request, Unauthorized, err.Error())
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
	_, err = s.StartWebSocket(response, request, lobby, user)
	if err != nil {
		_ = RejectConnection(response, request, InternalServerError, "cannot start websocket connection")
		return
	}
}

func (s *APIServer) connectLobby(response http.ResponseWriter, request *http.Request) {
	user, err := s.GetUser(request)
	if err != nil {
		_ = RejectConnection(response, request, Unauthorized, err.Error())
		return
	}
	lobbyId := mux.Vars(request)["lobby"]
	lobby, ok := s.Lobbies[lobbyId]
	if !ok {
		_ = RejectConnection(response, request, NotFound, "lobby not found")
		return
	}
	if user := lobby.GetUser(user); user == nil {
		_ = RejectConnection(response, request, Forbidden, "user not in lobby")
		return
	}
	_, err = s.StartWebSocket(response, request, lobby, user)
	if err != nil {
		_ = RejectConnection(response, request, InternalServerError, "cannot start websocket connection")
		return
	}
}

func (s *APIServer) getAllLobbies(response http.ResponseWriter, _ *http.Request) {
	type lobbyListType struct {
		Id         string `json:"id"`
		Owner      *User  `json:"owner"`
		Users      int    `json:"users"`
		MaxPlayers int    `json:"max_players"`
		State      any    `json:"state"`
	}

	lobbyList := make([]lobbyListType, 0, len(s.Lobbies))

	for key, lobby := range s.Lobbies {
		lobbyList = append(lobbyList, lobbyListType{
			Id:         key,
			Owner:      lobby.Owner,
			Users:      len(lobby.Users),
			MaxPlayers: lobby.Settings.MaxPlayers,
			State:      GetStateType(lobby.State),
		})
	}

	response.Header().Add("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err := json.NewEncoder(response).Encode(lobbyList)
	if err != nil {
		_ = fmt.Errorf("[API] error with encoding lobbies into json: %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *APIServer) GetUser(request *http.Request) (*User, error) {
	cookie, err := request.Cookie("jwt")
	if err != nil {
		return nil, errors.New("missing jwt cookie")
	}

	verifyTokenResponse, err := s.verifyJwt(cookie.Value)
	if err != nil {
		return nil, err
	}

	return &User{
		Id:             UserId(verifyTokenResponse.Id),
		Username:       verifyTokenResponse.Username,
		Email:          verifyTokenResponse.Email,
		Avatar:         verifyTokenResponse.Avatar,
		Role:           verifyTokenResponse.Role,
		Token:          cookie.Value,
		TokenExpiresAt: verifyTokenResponse.ExpiresAt,
	}, nil
}

func (s *APIServer) verifyJwt(jwt string) (*VerifyTokenResponse, error) {
	requestURL := fmt.Sprintf("%s/v1/validateToken", s.Config.BackendURL)
	requestBodyMap := map[string]string{"token": jwt}
	verifyTokenResponse := &VerifyTokenResponse{}

	err := utils.HttpPost(requestURL, map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", s.Config.BackendApiKey),
	}, requestBodyMap, verifyTokenResponse)

	log.Printf("[API] verifyJwt response ID: %v", verifyTokenResponse.Id)
	log.Printf("                   Username: %v", verifyTokenResponse.Username)
	log.Printf("                      Email: %v", verifyTokenResponse.Email)
	log.Printf("                     Avatar: %v", verifyTokenResponse.Avatar)
	log.Printf("                       Role: %v", verifyTokenResponse.Role)
	log.Printf("                  ExpiresAt: %v", verifyTokenResponse.ExpiresAt)

	return verifyTokenResponse, err
}
