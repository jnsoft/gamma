package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"math/big"

	"golang.org/x/crypto/scrypt"
	"golang.org/x/crypto/sha3"
)

const (
	// number of bits in a big.Word
	wordBits = 32 << (uint64(^big.Word(0)) >> 63)
	// number of bytes in a big.Word
	wordBytes = wordBits / 8
)

type KeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
}

type KeyPairSerialized struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

func NewKeyPair(bits int) (KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return KeyPair{}, err
	}
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	if err != nil {
		return KeyPair{}, err
	}
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{PrivateKey: privateKeyBytes, PublicKey: publicKeyBytes}, nil
}

// asymmetric elliptic-curve ECDSA key
func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// Ethereum address
func PubKeyToAddress(p *ecdsa.PublicKey) []byte {
	if p == nil || p.X == nil || p.Y == nil {
		return nil
	}
	pubBytes := MarshalCurve(p.X, p.Y)
	return Keccak256(pubBytes[1:])[12:]
}

// Marshal converts a point into the form specified in section 4.3.6 of ANSI X9.62.
func MarshalCurve(x, y *big.Int) []byte {
	byteLen := (elliptic.P256().Params().BitSize + 7) >> 3
	ret := make([]byte, 1+2*byteLen)
	ret[0] = 4 // uncompressed point flag
	readBits(x, ret[1:1+byteLen])
	readBits(y, ret[1+byteLen:])
	return ret
}

// Unmarshal converts a point, serialised by Marshal, into an x, y pair. On error, x = nil.
func UnmarshalCurve(data []byte) (x, y *big.Int) {
	byteLen := (elliptic.P256().Params().BitSize + 7) >> 3
	if len(data) != 1+2*byteLen {
		return
	}
	if data[0] != 4 { // uncompressed form
		return
	}
	x = new(big.Int).SetBytes(data[1 : 1+byteLen])
	y = new(big.Int).SetBytes(data[1+byteLen:])
	return
}

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func Keccak256(data ...[]byte) []byte {
	b := make([]byte, 32)
	d := sha3.New256() // NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return b
}

func GenerateNonce() uint32 {
	res, _ := generateUint32()
	if res != 0 {
		return res
	}
	panic(errors.New("generate uint32 failed"))
	// rand.NewSource(time.Now().UTC().UnixNano()) // using math rand
	//return rand.Uint32()
}

func generateUint32() (uint32, error) {
	var b [4]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b[:]), nil
}

func AesEncrypt(arr []byte, password string) ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, arr, nil), nil
}

func AesDecrypt(arr []byte, password string) ([]byte, error) {
	salt := arr[:16]
	key, err := scrypt.Key([]byte(password), salt, 16384, 8, 1, 32)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := arr[16 : 16+gcm.NonceSize()]
	ciphertext := arr[16+gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// readBits encodes the absolute value of bigint as big-endian bytes. Callers
// must ensure that buf has enough space. If buf is too short the result will
// be incomplete.
func readBits(bigint *big.Int, buf []byte) {
	i := len(buf)
	for _, d := range bigint.Bits() {
		for j := 0; j < wordBytes && i > 0; j++ {
			i--
			buf[i] = byte(d)
			d >>= 8
		}
	}
}
