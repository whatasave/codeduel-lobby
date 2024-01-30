package codeduel

import (
	"fmt"
	"time"

	"github.com/xedom/codeduel-lobby/codeduel/utils"
)

type Lobby struct {
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
	Ready []UserId
}

func NewLobby(owner *User) Lobby {
	return Lobby{
		Owner: owner,
		Users: map[UserId]*User{owner.Id: owner},
		Settings: Settings{
			MaxPlayers:       8,
			GameDuration:     time.Minute * 15,
			AllowedLanguages: []string{"typescript", "python"},
		},
		State: PreLobbyState{
			Ready: []UserId{},
		},
	}
}

func (lobby *Lobby) CanJoin(user *User) bool {
	return len(lobby.Users) < lobby.Settings.MaxPlayers
}

func (lobby *Lobby) AddUser(user *User) {
	lobby.Users[user.Id] = user
}

func (lobby *Lobby) SetSettings(settings Settings) {
	lobby.Settings = settings
}

func (lobby *Lobby) SetState(user *User, state string) error {
	if lobbyState, ok := lobby.State.(PreLobbyState); !ok {
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
