package hexutil

import (
	"encoding/hex"
	"fmt"
)

// Decode decodes a hex string with 0x prefix.
func Decode(input string) ([]byte, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("empty hex string")
	}
	if !has0xPrefix(input) {
		return nil, fmt.Errorf("hex string without 0x prefix: '%v'", input)
	}
	b, err := hex.DecodeString(input[2:])
	return b, err
}

// MustDecode panics for invalid input.
func MustDecode(input string) []byte {
	res, err := Decode(input)
	if err != nil {
		panic(err)
	}
	return res
}

// Encode encodes bs as a hex string with 0x prefix.
func Encode(bs []byte) string {
	enc := make([]byte, len(bs)*2+2)
	copy(enc, "0x")
	hex.Encode(enc[2:], bs)
	return string(enc)
}

func has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}
