package address

import (
	"encoding/hex"
	"fmt"

	"github.com/jnsoft/gamma/src/pkg/crypto"
)

const (
	// HashLength is the expected length of the hash
	HashLength = 32
	// AddressLength is the expected length of the address
	AddressLength = 20
)

type Address [AddressLength]byte

func PublicKeyToAddress(pub []byte) Address {
	hash := crypto.Sha3_256(pub)
	return Address(hash[HashLength-AddressLength:])
}

func (a Address) String() string {
	return hex.EncodeToString(a[:])
}

func (a Address) Hex() string {
	return hex.EncodeToString(a[:])
}

func ToAddress(b []byte) (Address, error) {
	if len(b) != AddressLength {
		return Address{}, fmt.Errorf("invalid address length")
	}
	var addr Address
	copy(addr[:], b)
	return addr, nil
}
