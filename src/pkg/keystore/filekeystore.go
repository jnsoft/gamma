package keystore

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"os"

	"github.com/jnsoft/gamma/src/pkg/crypto"
	"github.com/jnsoft/gamma/src/pkg/domain/address"
)

type FileKeystore struct {
	EncSeed []byte
	Nonce   []byte
	Salt    []byte
	seed    []byte // decrypted seed cached after Unlock; not serialized
}

func NewFileKeystore(password []byte) (*FileKeystore, error) {
	seed := make([]byte, 32)
	rand.Read(seed)

	salt := make([]byte, 16)
	rand.Read(salt)

	key, _ := crypto.DeriveKey(password, salt)
	nonce, enc, _ := crypto.Encrypt(key, seed)

	// zeroize local copies
	for i := range seed {
		seed[i] = 0
	}
	for i := range key {
		key[i] = 0
	}

	return &FileKeystore{
		EncSeed: enc,
		Nonce:   nonce,
		Salt:    salt,
	}, nil
}

func NewFileKeystoreFromSeed(seed, password []byte) (*FileKeystore, error) {
	if len(seed) == 0 {
		return nil, errors.New("empty seed")
	}
	salt := make([]byte, 16)
	rand.Read(salt)
	key, err := crypto.DeriveKey(password, salt)
	if err != nil {
		return nil, err
	}
	nonce, enc, err := crypto.Encrypt(key, seed)
	if err != nil {
		return nil, err
	}

	// zeroize local key
	for i := range key {
		key[i] = 0
	}

	return &FileKeystore{
		EncSeed: enc,
		Nonce:   nonce,
		Salt:    salt,
	}, nil
}

func (ks *FileKeystore) IsUnlocked() bool {
	return len(ks.seed) != 0
}

func (ks *FileKeystore) Unlock(password []byte) error {
	key, err := crypto.DeriveKey(password, ks.Salt)
	if err != nil {
		return err
	}
	plain, err := crypto.Decrypt(key, ks.Nonce, ks.EncSeed)
	// zeroize derived key ASAP
	for i := range key {
		key[i] = 0
	}

	if err != nil {
		return err
	}

	// cache in memory for subsequent derivations/signing
	ks.seed = make([]byte, len(plain))
	copy(ks.seed, plain)
	// zeroize temporary plaintext buffer
	for i := range plain {
		plain[i] = 0
	}

	return nil
}

func (ks *FileKeystore) Lock() {
	if !ks.IsUnlocked() {
		return
	}
	for i := range ks.seed {
		ks.seed[i] = 0
	}
	ks.seed = nil
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

func (ks *FileKeystore) ChangePassword(oldPwd, newPwd []byte) error {
	// verify old password by decrypting
	keyOld, err := crypto.DeriveKey(oldPwd, ks.Salt)
	if err != nil {
		return err
	}
	plain, err := crypto.Decrypt(keyOld, ks.Nonce, ks.EncSeed)
	// zeroize old key regardless of decrypt result
	for i := range keyOld {
		keyOld[i] = 0
	}
	if err != nil {
		return errors.New("invalid password")
	}
	// re-encrypt with new password; keep same salt or generate a new one
	newSalt := make([]byte, len(ks.Salt))
	copy(newSalt, ks.Salt)
	keyNew, err := crypto.DeriveKey(newPwd, newSalt)
	if err != nil {
		// zeroize plaintext on failure
		for i := range plain {
			plain[i] = 0
		}
		return err
	}
	nonce, enc, err := crypto.Encrypt(keyNew, plain)
	// zeroize new key and plaintext
	for i := range keyNew {
		keyNew[i] = 0
	}
	if err != nil {
		return err
	}
	ks.Salt = newSalt
	ks.Nonce = nonce
	ks.EncSeed = enc
	// update cached seed
	ks.seed = make([]byte, len(plain))
	copy(ks.seed, plain)
	for i := range plain {
		plain[i] = 0
	}
	return nil
}

func (ks *FileKeystore) ExportSeed() ([]byte, error) {
	if !ks.IsUnlocked() {
		return nil, errors.New("keystore locked")
	}
	out := make([]byte, len(ks.seed))
	copy(out, ks.seed)
	return out, nil
}

func AccountPath(coinType, account uint32) DerivationPath {
	return DerivationPath{
		Hardened(44),       // purpose
		Hardened(coinType), // coin type
		Hardened(account),  // account
	}
}

func ReceivingPath(coinType, account, change, index uint32) DerivationPath {
	base := AccountPath(coinType, account)
	return append(base, change, index)
}

func (ks *FileKeystore) DeriveAddress(path DerivationPath) (address.Address, error) {
	if !ks.IsUnlocked() {
		return address.Address{}, errors.New("keystore locked: call Unlock before deriving")
	}

	d, err := crypto.DeriveScalarFromSeed(ks.seed, path)
	if err != nil {
		return address.Address{}, err
	}
	priv := crypto.ECDSAP256FromScalar(d)

	// Uncompressed public key bytes (X||Y) consistent with address.PublicKeyToAddress
	pub := append(priv.PublicKey.X.Bytes(), priv.PublicKey.Y.Bytes()...)
	return address.PublicKeyToAddress(pub), nil

}

func (ks *FileKeystore) Sign(path DerivationPath, msg []byte) ([]byte, error) {
	if !ks.IsUnlocked() {
		return nil, errors.New("keystore locked: call Unlock before signing")
	}
	if len(msg) == 0 {
		return nil, errors.New("empty message")
	}

	d, err := crypto.DeriveScalarFromSeed(ks.seed, path)
	if err != nil {
		return nil, err
	}
	priv := crypto.ECDSAP256FromScalar(d)

	sig, err := crypto.SignP256ASN1(priv, msg)
	zeroKey(priv)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

var _ Keystore = (*FileKeystore)(nil)
