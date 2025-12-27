package address

import "github.com/jnsoft/gamma/src/pkg/crypto"

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
