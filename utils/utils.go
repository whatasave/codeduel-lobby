package utils

import (
	"encoding/json"
	"time"

	"github.com/xedom/codeduel-lobby/types"
)

func UnixTimeToTime(unixTime int64) time.Time {
	return time.Unix(unixTime, 0)
}

func PaserseMessage(message []byte) (types.ClientMessage, error) {
	var clientMessage types.ClientMessage
	err := json.Unmarshal(message, &clientMessage)
	if err != nil {
		return types.ClientMessage{}, err
	}
	return clientMessage, nil
}
