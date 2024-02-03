package codeduel

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

func UnmarshalPacket(message []byte, v any) error {
	var packetType struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(message, &packetType); err != nil {
		return err
	}

	var packet interface{}
	switch packetType.Type {
	case "updateSettings":
		packet = new(PacketInSettings)
	case "updatePlayerStatus":
		packet = new(PacketInUserStatus)
	case "startLobby":
		packet = new(PacketInStartLobby)
	default:
		return fmt.Errorf("Unknown message type: %s", packetType.Type)
	}

	if err := json.Unmarshal(message, &packet); err != nil {
		return err
	}

	return nil
}

func MarshalPacket(packet any) (any, error) {
	var packetType string
	switch packet.(type) {
	case PacketOutLobby:
		packetType = "lobby"
	default:
		return nil, fmt.Errorf("Unknown packet: %T", packet)
	}
	var m map[string]any
	p, err := json.Marshal(packet)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(p, &m)
	if err != nil {
		return nil, err
	}
	m["type"] = packetType
	return m, nil
}

func ReadPacket(connection *websocket.Conn, packet any) error {
	_, bytes, err := connection.ReadMessage()
	if err != nil {
		return err
	}
	return UnmarshalPacket(bytes, packet)
}

func SendPacket(connection *websocket.Conn, packet any) error {
	packet, err := MarshalPacket(packet)
	if err != nil {
		return err
	}
	return connection.WriteJSON(packet)
}

const (
	StatusReady    = "ready"
	StatusNotReady = "not_ready"
)

type PacketInSettings struct {
	Settings Settings
}

type PacketInUserStatus struct {
	Status string `json:"status"`
}

type PacketInStartLobby struct {
	Start bool `json:"start"`
}

type PacketOutLobby struct {
	Settings Settings         `json:"settings"`
	Users    map[UserId]*User `json:"users"`
	State    any              `json:"state"`
}
