package security

import (
	"math/rand"
	"time"
)

func GenerateNonce() uint32 {
	rand.NewSource(time.Now().UTC().UnixNano())
	return rand.Uint32()
}
