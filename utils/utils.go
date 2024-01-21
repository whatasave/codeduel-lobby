package utils

import (
	"time"
)

func UnixTimeToTime(unixTime int64) time.Time {
	return time.Unix(unixTime, 0)
}
