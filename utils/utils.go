package utils

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/xedom/codeduel-lobby/types"
)

func UnixTimeToTime(unixTime int64) time.Time {
	return time.Unix(unixTime, 0)
}

func ParseMessage(message []byte) (any, error) {
	var packet types.PacketIn
	// var clientMessage interface{}
	
	if err := json.Unmarshal(message, &packet); err != nil {
		return nil, err
	}

	var packetIn interface{}
	switch packet.Type {
	case "updateSettings":
		packetIn = new(types.SettingsPacketIn)
	case "updatePlayerStatus":
		packetIn = new(types.PlayerStatusPacketIn)
	case "startLobby":
		packetIn = new(types.StartLobbyPacketIn)
	default:
		return nil, fmt.Errorf("Unknown message type: %s", packet.Type)
	}
	
	if err := json.Unmarshal(message, &packetIn); err != nil {
		return nil, err
	}

	return packetIn, nil
}
