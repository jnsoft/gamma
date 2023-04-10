package misc

import (
	"time"
)

func GetTime() uint64 {
	return uint64(time.Now().Unix())
}
