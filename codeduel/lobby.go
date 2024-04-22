package codeduel

import (
	"context"
	"fmt"
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
	MaxPlayers       int           `json:"maxPlayers"`
	GameDuration     time.Duration `json:"gameDuration"`
	AllowedLanguages []string      `json:"allowedLanguages"`
}

type PreLobbyState struct {
	Type  string   `json:"type"`
	Ready []UserId `json:"ready"`
}

type GameLobbyState struct {
	Type       string                        `json:"type"`
	Challenge  Challenge                     `json:"challenge"`
	StartTime  time.Time                     `json:"startTime"`
	UsersState map[UserId]UserGameLobbyState `json:"usersState"`
	context    context.CancelCauseFunc
}

type UserGameLobbyState struct {
	LastRunResult *RunResult `json:"lastRunResult"`
	SubmitResult  *RunResult `json:"submitResult"`
}

type RunResult struct {
	Results []ExecutionResult `json:"results"`
	Date    time.Time         `json:"date"`
}

type Challenge struct {
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	TestCases       []TestCase `json:"testCases"`
	HiddenTestCases []TestCase `json:"hiddenTestCases"`
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
			GameDuration:     time.Minute * 15,
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

func (lobby *Lobby) AddUser(user *User) {
	fmt.Printf("Adding user to lobby: %v\n", user)
	lobby.Users[user.Id] = user
}

func (lobby *Lobby) SetSettings(settings Settings) {
	lobby.Settings = settings
}

func (lobby *Lobby) SetState(user *User, state string) error {
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

func (lobby *Lobby) RunTest(user *User, runner *Runner, language string, code string) ([]ExecutionResult, error) {
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
	state.UsersState[user.Id] = UserGameLobbyState{
		LastRunResult: &RunResult{
			Results: result,
			Date:    time.Now(),
		},
		SubmitResult: state.UsersState[user.Id].SubmitResult,
	}
	return result, nil
}

func (lobby *Lobby) Submit(user *User, runner *Runner, language string, code string) ([]ExecutionResult, error) {
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
	state.UsersState[user.Id] = UserGameLobbyState{
		LastRunResult: state.UsersState[user.Id].LastRunResult,
		SubmitResult: &RunResult{
			Results: result,
			Date:    time.Now(),
		},
	}
	return result, nil
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
	lobby.State = GameLobbyState{
		Type:      "game",
		Challenge: RandomChallenge(),
		StartTime: time.Now(),
		context:   cancel,
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
	utils.WaitUntil(ctx, state.StartTime.Add(lobby.Settings.GameDuration))
	delete(s.Lobbies, lobby.Id)
}
