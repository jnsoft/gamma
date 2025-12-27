package keystore

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/google/uuid"
	"github.com/jnsoft/gamma/src/pkg/crypto"
	"github.com/jnsoft/gamma/src/pkg/domain/address"
)

type DerivationPath []uint32

type Keystore interface {
	DeriveAddress(path DerivationPath) (address.Address, error)
	Sign(path DerivationPath, msg []byte) ([]byte, error)
}

func (ks *KeyStore) NewKey() (*Key, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	key := &Key{
		Id:         uuid.New(),
		Address:    address.PublicKeyToAddress(crypto.FromECDSAPub(&privKey.PublicKey)),
		PrivateKey: privKey,
	}

	return key, nil
}

type KeyStore struct {
}

// zeroKey zeroes a private key in memory.
func zeroKey(k *ecdsa.PrivateKey) {
	b := k.D.Bits()
	clear(b)
}
