package hexutil

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

const badNibble = ^uint64(0)

var (
	bytesT  = reflect.TypeOf(Bytes(nil))
	uint64T = reflect.TypeOf(Uint64(0))
)

type Bytes []byte

type decError struct{ msg string }

// Uint64 marshals/unmarshals as a JSON string with 0x prefix.
// The zero value marshals as "0x0".
type Uint64 uint64

func (err decError) Error() string { return err.msg }

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

func bytesHave0xPrefix(input []byte) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}

func checkNumberText(input []byte) (raw []byte, err error) {
	if len(input) == 0 {
		return nil, nil // empty strings are allowed
	}
	if !bytesHave0xPrefix(input) {
		return nil, errors.New("hex string without 0x prefix")
	}
	input = input[2:]
	if len(input) == 0 {
		return nil, errors.New("hex string \"0x\"")
	}
	if len(input) > 1 && input[0] == '0' {
		return nil, errors.New("hex number with leading zero digits")
	}
	return input, nil
}

// JSON
// Bytes marshals/unmarshals as a JSON string with 0x prefix.
// The empty slice marshals as "0x".

// MarshalText implements encoding.TextMarshaler
func (b Bytes) MarshalText() ([]byte, error) {
	res := make([]byte, len(b)*2+2)
	copy(res, `0x`)
	hex.Encode(res[2:], b)
	return res, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Bytes) UnmarshalJSON(input []byte) error {
	if !isString(input) {
		return errNonString(bytesT)
	}
	return wrapTypeError(b.UnmarshalText(input[1:len(input)-1]), bytesT)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *Bytes) UnmarshalText(input []byte) error {
	raw, err := checkText(input, true)
	if err != nil {
		return err
	}
	dec := make([]byte, len(raw)/2)
	if _, err = hex.Decode(dec, raw); err != nil {
		// err = mapError(err)
	} else {
		*b = dec
	}
	return err
}

func UnmarshalFixedJSON(typ reflect.Type, input, out []byte) error {
	if !isString(input) {
		return errNonString(typ)
	}
	return wrapTypeError(UnmarshalFixedText(typ.String(), input[1:len(input)-1], out), typ)
}

func UnmarshalFixedText(typname string, input, out []byte) error {
	raw, err := checkText(input, true)
	if err != nil {
		return err
	}
	if len(raw)/2 != len(out) {
		return fmt.Errorf("hex string has length %d, want %d for %s", len(raw), len(out)*2, typname)
	}
	// Pre-verify syntax before modifying out.
	for _, b := range raw {
		if decodeNibble(b) == badNibble {
			return fmt.Errorf("invalid hex string")
		}
	}
	hex.Decode(out, raw)
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler
func (b *Uint64) UnmarshalText(input []byte) error {
	raw, err := checkNumberText(input)
	if err != nil {
		return err
	}
	if len(raw) > 16 {
		return fmt.Errorf("hex number > 64 bits")
	}
	var dec uint64
	for _, byte := range raw {
		nib := decodeNibble(byte)
		if nib == badNibble {
			return fmt.Errorf("invalid hex string")
		}
		dec *= 16
		dec += nib
	}
	*b = Uint64(dec)
	return nil
}

func isString(input []byte) bool {
	return len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"'
}

func checkText(input []byte, wantPrefix bool) ([]byte, error) {
	if len(input) == 0 {
		return nil, nil // empty strings are allowed
	}
	if bytesHave0xPrefix(input) {
		input = input[2:]
	} else if wantPrefix {
		return nil, fmt.Errorf("hex string without 0x prefix")
	}
	if len(input)%2 != 0 {
		return nil, fmt.Errorf("hex string of odd length")
	}
	return input, nil
}

func decodeNibble(in byte) uint64 {
	switch {
	case in >= '0' && in <= '9':
		return uint64(in - '0')
	case in >= 'A' && in <= 'F':
		return uint64(in - 'A' + 10)
	case in >= 'a' && in <= 'f':
		return uint64(in - 'a' + 10)
	default:
		return badNibble
	}
}

func wrapTypeError(err error, typ reflect.Type) error {
	if _, ok := err.(*decError); ok {
		return &json.UnmarshalTypeError{Value: err.Error(), Type: typ}
	}
	return err
}

func errNonString(typ reflect.Type) error {
	return &json.UnmarshalTypeError{Value: "non-string", Type: typ}
}
