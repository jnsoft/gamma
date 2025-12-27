package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
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
