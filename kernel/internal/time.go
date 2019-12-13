package internal

import "time"

func TimeToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / 1000000
}
