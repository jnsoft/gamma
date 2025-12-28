package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

// GeneratePrivateKey returns a new ECDSA P-256 private key.
func GeneratePrivateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader) // replace w/ secp256k1
}

// FromECDSAPub encodes an ECDSA public key in uncompressed form: 0x04 || X || Y.
func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	curve := pub.Curve
	byteLen := (curve.Params().BitSize + 7) / 8

	x := pub.X.Bytes()
	y := pub.Y.Bytes()

	enc := make([]byte, 1+2*byteLen)
	enc[0] = 0x04
	// left-pad X and Y to fixed length
	copy(enc[1+byteLen-len(x):1+byteLen], x)
	copy(enc[1+2*byteLen-len(y):1+2*byteLen], y)
	return enc
}

// parse pub: 0x04 || X || Y
func ToECDSAPub(pub []byte) (*ecdsa.PublicKey, error) {
	p := elliptic.P256().Params()
	byteLen := (p.BitSize + 7) / 8
	if len(pub) != 1+2*byteLen || pub[0] != 0x04 {
		return nil, errors.New("invalid pubkey format")
	}
	x := new(big.Int).SetBytes(pub[1 : 1+byteLen])
	y := new(big.Int).SetBytes(pub[1+byteLen : 1+2*byteLen])
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
}

func SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error) {
	// The Go standard library does not provide ECDSA public-key recovery
	// from a signature (r, s) and a message hash. If you need public key
	// recovery (as used by Ethereum/secp256k1), use a dedicated library
	// such as github.com/ethereum/go-ethereum/crypto or
	// github.com/btcsuite/btcd/btcec which expose RecoverPubkey-like APIs.
	return nil, fmt.Errorf("ecdsa public key recovery not implemented; use a secp256k1/ethereum crypto library")
}
