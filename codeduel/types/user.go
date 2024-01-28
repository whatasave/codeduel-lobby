package types

import "github.com/xedom/codeduel-lobby/packets"

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
	Type       string                   `json:"type"`
	Status     string                   `json:"status"`
	StartLobby bool                     `json:"startLobby"`
	Code       string                   `json:"code"`
	Settings   packets.SettingsPacketIn `json:"settings"`
}
