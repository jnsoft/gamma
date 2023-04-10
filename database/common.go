package database

import (
	"bytes"
	"encoding/hex"
	"reflect"

	"github.com/jnsoft/gamma/util/hexutil"
	"golang.org/x/crypto/sha3"
)

// Lengths of hashes and addresses in bytes.
const (
	// HashLength is the expected length of the hash
	HashLength = 32
	// AddressLength is the expected length of the address
	AddressLength = 20

	A0 string = "0x0000000000000000000000000000000000000009"
	A1 string = "0x0000000000000000000000000000000000000001"
	A2 string = "0x0000000000000000000000000000000000000002"
	A3 string = "0x0000000000000000000000000000000000000003"
)

var (
	hashT    = reflect.TypeOf(Hash{})
	addressT = reflect.TypeOf(Address{})
)

type Hash [HashLength]byte

type Address [AddressLength]byte

////////////////// HASH //////////////

func (h Hash) Hex() string {
	return hex.EncodeToString(h[:])
}

func (h Hash) IsEmpty() bool {
	emptyHash := Hash{}
	return bytes.Equal(emptyHash[:], h[:])
}

func (h Hash) MarshalText() ([]byte, error) {
	return []byte(h.Hex()), nil
}

func (h *Hash) UnmarshalText(data []byte) error {
	_, err := hex.Decode(h[:], data)
	return err
}

////////////////// ADDRESS //////////////

func ToAddress(hex string) Address {
	return Address(hexutil.MustDecode(hex))
}

// String implements fmt.Stringer.
func (a Address) String() string {
	return a.Hex()
}

// Hex returns an EIP55-compliant hex string representation of the address.
func (a Address) Hex() string {
	return string(a.checksumHex())
}

func (a *Address) checksumHex() []byte {
	buf := a.hex()

	// compute checksum
	sha := sha3.NewLegacyKeccak256()
	sha.Write(buf[2:])
	hash := sha.Sum(nil)
	for i := 2; i < len(buf); i++ {
		hashByte := hash[(i-2)/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if buf[i] > '9' && hashByte > 7 {
			buf[i] -= 32
		}
	}
	return buf[:]
}

func (a Address) hex() []byte {
	var buf [len(a)*2 + 2]byte
	copy(buf[:2], "0x")
	hex.Encode(buf[2:], a[:])
	return buf[:]
}

func (a Address) MarshalText() ([]byte, error) {
	return []byte(a.Hex()), nil
}

// UnmarshalJSON parses a hash in hex syntax.
func (a *Address) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(addressT, input, a[:])
}

func (v *Address) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("MyType", input, v[:])
}
