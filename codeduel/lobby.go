package codeduel

import "time"

type Lobby struct {
	Owner    *User
	Users    []*User
	Settings Settings
}

type Settings struct {
	MaxPlayers       int           `json:"max_players"`
	GameDuration     time.Duration `json:"game_duration"`
	AllowedLanguages []string      `json:"allowed_languages"`
}

func NewLobby(owner *User) Lobby {
	return Lobby{
		Owner: owner,
		Users: []*User{owner},
		Settings: Settings{
			MaxPlayers:       8,
			GameDuration:     time.Minute * 15,
			AllowedLanguages: []string{"typescript", "python"},
		},
	}
}

func (lobby *Lobby) CanJoin(user *User) bool {
	return len(lobby.Users) < lobby.Settings.MaxPlayers
}

func (lobby *Lobby) AddUser(user *User) {
	lobby.Users = append(lobby.Users, user)
}

func (lobby *Lobby) SetSettings(settings Settings) {
	lobby.Settings = settings
}
