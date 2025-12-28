package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"math/big"

	"golang.org/x/crypto/hkdf"
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

// DeriveScalarFromSeed computes d = HMAC-SHA256(seed, ser(path)) mod N, with HKDF fallback if zero.
func DeriveScalarFromSeed(seed []byte, path []uint32) (*big.Int, error) {
	if len(seed) == 0 {
		return nil, errors.New("empty seed")
	}
	ser := serializePath(path)

	mac := hmac.New(sha256.New, seed)
	mac.Write(ser)
	keyMaterial := mac.Sum(nil) // 32 bytes

	curve := elliptic.P256()
	N := curve.Params().N

	d := new(big.Int).SetBytes(keyMaterial)
	d.Mod(d, N)
	if d.Sign() == 0 {
		h := hkdf.New(sha256.New, seed, ser, []byte("gamma-hd-fallback"))
		buf := make([]byte, 32)
		if _, err := io.ReadFull(h, buf); err != nil {
			return nil, err
		}
		d.SetBytes(buf)
		d.Mod(d, N)
		if d.Sign() == 0 {
			return nil, errors.New("derived zero scalar")
		}
	}
	return d, nil
}

// ECDSAP256FromScalar constructs an ECDSA P-256 private key from scalar d.
func ECDSAP256FromScalar(d *big.Int) *ecdsa.PrivateKey {
	curve := elliptic.P256()
	priv := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{Curve: curve},
		D:         d,
	}
	priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(d.Bytes())
	return priv
}

// SignP256ASN1 signs msg using ECDSA P-256 with DER-encoded output.
func SignP256ASN1(priv *ecdsa.PrivateKey, msg []byte) ([]byte, error) {
	return ecdsa.SignASN1(rand.Reader, priv, msg)
}

// Verify verifies ASN.1 (r,s) over hash using uncompressed pubkey 0x04||X||Y,
func Verify(msg, sig, pub []byte) error {

	pubKey, err := ToECDSAPub(pub)
	if err != nil {
		return err
	}

	ok := ecdsa.VerifyASN1(pubKey, msg, sig)
	if !ok {
		return errors.New("signature verify failed")
	}

	return nil
}

// SerializePath encodes a derivation path as big-endian uint32 bytes.
func serializePath(path []uint32) []byte {
	ser := make([]byte, 0, 4*len(path))
	for _, idx := range path {
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], idx)
		ser = append(ser, b[:]...)
	}
	return ser
}

/* DONT WANT TO USE ADDRESS HERE...
// derive address and compare
    pubXY := append(x.Bytes(), y.Bytes()...)
    got := address.PublicKeyToAddress(pubXY)
    if got != want {
        return errors.New("pubkey does not match address")
    }
    return nil
*/
