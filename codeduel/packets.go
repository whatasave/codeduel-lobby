package codeduel

import (
	"encoding/json"
	"fmt"
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
