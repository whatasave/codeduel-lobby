package codeduel

import (
	"encoding/json"
	"fmt"
	"time"

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
	case "check":
		typedPacket = new(PacketInCheck)
	case "submit":
		typedPacket = new(PacketInSubmit)
	default:
		return fmt.Errorf("unknown message type: %s", packetType.Type)
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
	case PacketOutGameStarted:
		packetType = "gameStarted"
	case PacketOutCheckResult:
		packetType = "checkResult"
	default:
		return nil, fmt.Errorf("unknown packet: %T", packet)
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

func (lobby *Lobby) BroadcastPacket(packet any) []User {
	users := make([]User, 0, len(lobby.Users))
	for _, user := range lobby.Users {
		if user.Connection != nil {
			err := SendPacket(user.Connection, packet)
			if err != nil {
				fmt.Printf("error while sending packet to user %v: %v\n", user, err)
				users = append(users, *user)
			}
		} else {
			users = append(users, *user)
		}
	}
	return users
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

type PacketInCheck struct {
	Code string `json:"code"`
}

type PacketInSubmit struct {
	Code string `json:"code"`
}

type PacketOutLobby struct {
	LobbyID  string           `json:"id"`
	Settings Settings         `json:"settings"`
	Owner    *User            `json:"owner"`
	Users    map[UserId]*User `json:"users"`
	State    any              `json:"state"`
}

type PacketOutGameStarted struct {
	StartTime time.Time `json:"startTime"`
	Challenge Challenge `json:"challenge"`
}

type PacketOutCheckResult struct {
	Result []ExecutionResult `json:"result"`
}
