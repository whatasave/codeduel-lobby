package codeduel

import (
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
	MaxPlayers       int           `json:"max_players"`
	GameDuration     time.Duration `json:"game_duration"`
	AllowedLanguages []string      `json:"allowed_languages"`
}

type PreLobbyState struct {
	Type  string   `json:"type"`
	Ready []UserId `json:"ready"`
}

type GameLobbyState struct {
	Type string `json:"type"`
}

func NewLobby(owner *User) Lobby {
	return Lobby{
		Id:    uuid.NewString(),
		Owner: owner,
		Users: map[UserId]*User{owner.Id: owner},
		Settings: Settings{
			MaxPlayers:       8,
			GameDuration:     time.Minute * 15,
			AllowedLanguages: []string{"typescript", "python"},
		},
		State: PreLobbyState{
			Type:  "pre_lobby",
			Ready: []UserId{},
		},
	}
}

func (lobby *Lobby) CannotJoin(user *User) error {
	if _, ok := lobby.State.(PreLobbyState); !ok {
		return fmt.Errorf("Lobby is not in PreLobby")
	}
	if len(lobby.Users) >= lobby.Settings.MaxPlayers {
		return fmt.Errorf("Lobby is full")
	}
	return nil
}

func (lobby *Lobby) GetUser(user *User) *User {
	return lobby.Users[user.Id]
}

func (lobby *Lobby) AddUser(user *User) {
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
			return fmt.Errorf("Unknown user state: %v", state)
		}
		return nil
	} else {
		return fmt.Errorf("Lobby is not in PreLobby")
	}
}

func (lobby *Lobby) Start() error {
	if _, ok := lobby.State.(PreLobbyState); ok {
		lobby.State = GameLobbyState{
			Type: "game",
		}
		return nil
	} else {
		return fmt.Errorf("Lobby is not in PreLobby")
	}
}
