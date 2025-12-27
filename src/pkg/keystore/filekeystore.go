package keystore

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/json"
	"os"

	"github.com/jnsoft/gamma/src/crypto"
)

type FileKeystore struct {
	EncSeed []byte
	Nonce   []byte
	Salt    []byte
}

func NewFileKeystore(password []byte) (*FileKeystore, error) {
	seed := make([]byte, 32)
	rand.Read(seed)

	salt := make([]byte, 16)
	rand.Read(salt)

	key, _ := crypto.DeriveKey(password, salt)
	nonce, enc, _ := crypto.Encrypt(key, seed)

	return &FileKeystore{
		EncSeed: enc,
		Nonce:   nonce,
		Salt:    salt,
	}, nil
}

func (ks *FileKeystore) Unlock(password []byte) ([]byte, error) {
	key, _ := crypto.DeriveKey(password, ks.Salt)
	return crypto.Decrypt(key, ks.Nonce, ks.EncSeed)
}

func (ks *FileKeystore) Save(path string) error {
	data, _ := json.MarshalIndent(ks, "", "  ")
	return os.WriteFile(path, data, 0600)
}

func LoadFileKeystore(path string) (*FileKeystore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ks FileKeystore
	if err := json.Unmarshal(data, &ks); err != nil {
		return nil, err
	}
	return &ks, nil
}

func (ks *FileKeystore) DeriveAddress(path DerivationPath) (types.Address, error) {
	// Placeholder derivation, use BIP-32 derivation
	priv, _ := crypto.GeneratePrivateKey()
	pub := append(priv.PublicKey.X.Bytes(), priv.PublicKey.Y.Bytes()...)
	return crypto.PublicKeyToAddress(pub), nil
}

func (ks *FileKeystore) Sign(path DerivationPath, msg []byte) ([]byte, error) {
	priv, _ := crypto.GeneratePrivateKey()
	r, s, _ := ecdsa.Sign(nil, priv, msg)
	return append(r.Bytes(), s.Bytes()...), nil
}
