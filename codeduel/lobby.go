package codeduel

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/xedom/codeduel-lobby/codeduel/utils"
)

type Lobby struct {
	Id       string
	Owner    *User
	Users    map[UserId]*User
	Settings Settings
	State    any
}

type Settings struct {
	Mode             string        `json:"mode"`
	MaxPlayers       int           `json:"maxPlayers"`
	GameDuration     time.Duration `json:"gameDuration"`
	AllowedLanguages []string      `json:"allowedLanguages"`
}

type PreLobbyState struct {
	Type  string   `json:"type"`
	Ready []UserId `json:"ready"`
}

type GameLobbyState struct {
	Type        string                        `json:"type"`
	Challenge   Challenge                     `json:"challenge"`
	StartTime   time.Time                     `json:"startTime"`
	UsersState  map[UserId]UserGameLobbyState `json:"usersState"`
	SubmitCount int                           `json:"submitCount"`
	context     context.CancelCauseFunc
}

type UserGameLobbyState struct {
	LastRunResult *RunResult `json:"lastRunResult"`
	SubmitResult  *RunResult `json:"submitResult"`
}

type RunResult struct {
	Code        string            `json:"code"`
	Language    string            `json:"language"`
	Results     []ExecutionResult `json:"results"`
	PassedTests int               `json:"passedTests"`
	Date        time.Time         `json:"date"`
}

type ChallengeId int32
type Challenge struct {
	Id    ChallengeId `json:"id"`
	Owner struct {
		Id       int    `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
	} `json:"owner"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"` // markdown maybe the link to the file

	TestCases       []TestCase `json:"testCases"`
	HiddenTestCases []TestCase `json:"hiddenTestCases"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type TestCase struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

func NewLobby(owner *User, allowedLanguages []string) Lobby {
	return Lobby{
		Id:    uuid.NewString(),
		Owner: owner,
		Users: map[UserId]*User{owner.Id: owner},
		Settings: Settings{
			MaxPlayers:       8,
			GameDuration:     time.Minute * 15 / time.Second,
			AllowedLanguages: allowedLanguages,
		},
		State: PreLobbyState{
			Type:  "preLobby",
			Ready: []UserId{},
		},
	}
}

func (lobby *Lobby) CannotJoin(_ *User) error {
	if _, ok := lobby.State.(PreLobbyState); !ok {
		return fmt.Errorf("lobby is not in PreLobby")
	}
	if len(lobby.Users) >= lobby.Settings.MaxPlayers {
		return fmt.Errorf("lobby is full")
	}
	return nil
}

func (lobby *Lobby) GetUser(user *User) *User {
	return lobby.Users[user.Id]
}

func (lobby *Lobby) GetReadyUsers() []UserId {
	if lobbyState, ok := lobby.State.(PreLobbyState); ok {
		return lobbyState.Ready
	}
	return []UserId{}
}

func (lobby *Lobby) AddUser(user *User) {
	log.Printf("Adding user to lobby: %v\n", user.Username)
	lobby.Users[user.Id] = user
}

func (lobby *Lobby) SetSettings(settings Settings) {
	lobby.Settings = settings
}

func (lobby *Lobby) SetReadyState(user *User, state string) error {
	if lobbyState, ok := lobby.State.(PreLobbyState); ok {
		if state == StatusReady {
			lobbyState.Ready = append(lobbyState.Ready, user.Id)
		} else if state == StatusNotReady {
			lobbyState.Ready = utils.Remove(lobbyState.Ready, user.Id)
		} else {
			return fmt.Errorf("unknown user state: %v", state)
		}
		return nil
	} else {
		return fmt.Errorf("lobby is not in PreLobby")
	}
}

func (lobby *Lobby) RunTest(user *User, runner *Runner, language string, code string) (*RunResult, error) {
	state, ok := lobby.State.(GameLobbyState)
	if !ok {
		return nil, fmt.Errorf("lobby is not in game state")
	}
	var input []string
	for _, testCase := range state.Challenge.TestCases {
		input = append(input, testCase.Input)
	}
	result, err := runner.Run(language, code, input)
	if err != nil {
		return nil, fmt.Errorf("error while running code: %v", err)
	}
	runResult := RunResult{
		Code:        code,
		Language:    language,
		Results:     result,
		Date:        time.Now(),
		PassedTests: testsPassed(state.Challenge.TestCases, result),
	}
	state.UsersState[user.Id] = UserGameLobbyState{
		LastRunResult: &runResult,
		SubmitResult:  state.UsersState[user.Id].SubmitResult,
	}
	return &runResult, nil
}

func (lobby *Lobby) Submit(user *User, runner *Runner, language string, code string) (*RunResult, error) {
	state, ok := lobby.State.(GameLobbyState)
	if !ok {
		return nil, fmt.Errorf("lobby is not in game state")
	}
	if state.UsersState[user.Id].SubmitResult != nil {
		return nil, fmt.Errorf("submit result is already set")
	}
	var input []string
	for _, testCase := range state.Challenge.HiddenTestCases {
		input = append(input, testCase.Input)
	}
	result, err := runner.Run(language, code, input)
	if err != nil {
		return nil, fmt.Errorf("error while running code: %v", err)
	}
	runResult := RunResult{
		Code:        code,
		Language:    language,
		Results:     result,
		Date:        time.Now(),
		PassedTests: testsPassed(state.Challenge.HiddenTestCases, result),
	}
	state.UsersState[user.Id] = UserGameLobbyState{
		LastRunResult: state.UsersState[user.Id].LastRunResult,
		SubmitResult:  &runResult,
	}
	state.SubmitCount++
	if state.SubmitCount == len(lobby.Users) {
		state.context(fmt.Errorf("all users submitted"))
	}
	return &runResult, nil
}

func (lobby *Lobby) KickUser(userId UserId) error {
	if _, ok := lobby.State.(PreLobbyState); ok {
		delete(lobby.Users, userId)
		return nil
	}

	return fmt.Errorf("lobby is not in PreLobby")
}

func testsPassed(testCases []TestCase, results []ExecutionResult) int {
	passed := 0
	for i, test := range results {
		if test.Error == "" && test.Status == 0 && test.Output == testCases[i].Output {
			passed++
		}
	}
	return passed
}

func GetStateType(state any) string {
	switch state.(type) {
	case PreLobbyState:
		return "preLobby"
	case GameLobbyState:
		return "game"
	default:
		return "unknown"
	}
}

func (s *APIServer) StartLobby(lobby *Lobby, ctx context.Context) error {
	if _, ok := lobby.State.(PreLobbyState); !ok {
		return fmt.Errorf("lobby is not in PreLobby")
	}
	ctx, cancel := context.WithCancelCause(ctx)

	randomChallenge, err := s.Backend.GetRandomChallenge()
	if err != nil {
		return fmt.Errorf("error while getting random challenge: %v", err)
	}

	lobby.State = GameLobbyState{
		Type:        "game",
		Challenge:   *randomChallenge,
		StartTime:   time.Now(),
		UsersState:  map[UserId]UserGameLobbyState{},
		SubmitCount: 0,
		context:     cancel,
	}
	go s.HandleGame(lobby, ctx)
	return nil
}

func (s *APIServer) HandleGame(lobby *Lobby, ctx context.Context) {
	state := lobby.State.(GameLobbyState)
	lobby.BroadcastPacket(PacketOutGameStarted{
		state.StartTime,
		state.Challenge,
	})
	err := s.Backend.CreateLobby(lobby)
	if err != nil {
		_ = fmt.Errorf("error while creating lobby: %v", err)
	}
	utils.WaitUntil(ctx, state.StartTime.Add(lobby.Settings.GameDuration))
	delete(s.Lobbies, lobby.Id)
	err = s.Backend.EndLobby(lobby)
	if err != nil {
		_ = fmt.Errorf("error while ending lobby: %v", err)
	}
}

func (s *APIServer) DeleteLobby(lobby *Lobby, ctx context.Context) error {
	if _, ok := lobby.State.(PreLobbyState); !ok {
		return fmt.Errorf("lobby is not in PreLobby")
	}

	lobby.BroadcastPacket(PacketOutLobbyDelete{
		Deleted: true,
	})

	delete(s.Lobbies, lobby.Id)
	return nil
}
