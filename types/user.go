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

type Settings struct {
	LockLobby    bool     `json:"lockLobby"`
	MaxPlayers   int      `json:"maxPlayers"`
	MaxDuration  int      `json:"maxDuration"`
	AllowedLangs []string `json:"allowedLangs"`
}
type ClientMessage struct {
	Status   string   `json:"status"`
	Code     string   `json:"code"`
	Settings Settings `json:"settings"`
}
