package types

const (
	NOT_READY = "NOT_READY"
	READY     = "READY"
	IN_MATCH  = "IN_MATCH"
	DONE      = "DONE"
)

const (
	STARTING = "STARTING"
	ONGOING  = "ONGOING"
	FINISHED = "FINISHED"
)

type ClientLobby struct {
	Type       string   `json:"type"`
	Status     string   `json:"status"`
	StartLobby bool     `json:"startLobby"`
	Code       string   `json:"code"`
	Settings   SettingsPacketIn `json:"settings"`
}

type PacketOut struct {}

type PacketIn struct {
	Type string `json:"type"`
}

type SettingsPacketIn struct {
	LockLobby    bool     `json:"lockLobby"`
	MaxPlayers   int      `json:"maxPlayers"`
	MaxDuration  int      `json:"maxDuration"`
	AllowedLangs []string `json:"allowedLangs"`
}

type PlayerStatusPacketIn struct {
	Status string `json:"status"`
}

type StartLobbyPacketIn struct {
	Start bool `json:"start"`
}
