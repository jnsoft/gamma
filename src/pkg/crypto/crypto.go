package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	"golang.org/x/crypto/scrypt"
)

func DeriveKey(password, salt []byte) ([]byte, error) {
	return scrypt.Key(password, salt, 1<<18, 8, 1, 32)
}

func Encrypt(key, plaintext []byte) (nonce, ciphertext []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	nonce = make([]byte, gcm.NonceSize())
	rand.Read(nonce)
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return
}

func Decrypt(key, nonce, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, ciphertext, nil)
}
