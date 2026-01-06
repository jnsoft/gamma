package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/sha3"
)

type Hash [32]byte

func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(h[:]))
}

func (h *Hash) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// Handle optional 0x prefix
	if len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		s = s[2:]
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	if len(b) != 32 {
		return fmt.Errorf("invalid hash length: need 32 bytes, got %d", len(b))
	}
	copy(h[:], b)
	return nil
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func BytesToHash(b []byte) Hash {
	var h Hash
	if len(b) > 32 {
		b = b[:32]
	}
	copy(h[32-len(b):], b)
	return h
}

func Keccak256(data []byte) Hash {
	h := sha3.NewLegacyKeccak256()
	h.Write(data)
	return BytesToHash(h.Sum(nil))
}

func Sha3_256(data []byte) Hash {
	h := sha3.New256()
	h.Write(data)
	return BytesToHash(h.Sum(nil))
}

func Sha2_256(data []byte) Hash {
	h := sha256.New()
	h.Write(data)
	return BytesToHash(h.Sum(nil))
}
