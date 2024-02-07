package codeduel

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

func UnmarshalPacket(message []byte, packet *any) error {
	var packetType struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(message, &packetType); err != nil {
		return err
	}

	var typedPacket any
	switch packetType.Type {
	case "updateSettings":
		typedPacket = new(PacketInSettings)
	case "updatePlayerStatus":
		typedPacket = new(PacketInUserStatus)
	case "startLobby":
		typedPacket = new(PacketInStartLobby)
	default:
		return fmt.Errorf("Unknown message type: %s", packetType.Type)
	}

	if err := json.Unmarshal(message, &typedPacket); err != nil {
		return err
	}

	*packet = typedPacket

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

func ReadPacket(connection *websocket.Conn, packet *any) error {
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
	LobbyID  string           `json:"id"`
	Settings Settings         `json:"settings"`
	Owner    *User            `json:"owner"`
	Users    map[UserId]*User `json:"users"`
	State    any              `json:"state"`
}
