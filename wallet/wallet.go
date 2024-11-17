package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/jnsoft/gamma/database"
	"github.com/jnsoft/gamma/util/security"
)

const keystoreDirName = "keystore"

const USE_RSA = true

type KeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
}

type KeyPairSerialized struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

// Wallet is a keypair with an ecrypted private key
type Wallet struct {
	Key KeyPair
	//PrivateKey string `json:"private_key"`
	//PublicKey  string `json:"public_key"`
}

func GetKeystoreDirPath(dataDir string) string {
	return filepath.Join(dataDir, keystoreDirName)
}

func NewKeyPair() KeyPair {
	return KeyPair{PublicKey: []byte{}, PrivateKey: []byte{}}
}

func NewWallet() Wallet {
	return Wallet{NewKeyPair()}
}

func CreateWallet(path, pwd string) (Wallet, error) {
	keypair, err := createKeyPair(path, pwd, USE_RSA)
	if err != nil {
		return NewWallet(), err
	}

	w := Wallet{Key: KeyPair{PublicKey: keypair.PublicKey, PrivateKey: keypair.PrivateKey}}

	if err := w.saveToFile(path, pwd); err != nil {
		return NewWallet(), err
	}
	return w, nil
}

func GetWallet(path, pwd string) (Wallet, error) {
	w, err := readWalletFromFile(path, pwd)
	if err != nil {
		return NewWallet(), err
	}
	return w, nil
}

func (w Wallet) Hex() string {
	return hex.EncodeToString(w.Key.PublicKey)
}

func (w Wallet) PrivateKeyString() string {
	return hex.EncodeToString(w.Key.PrivateKey)
}

func (w Wallet) SignTx(tx database.Tx, from database.Address, pwd string) {
	panic("not implemented")
}

func createKeyPair(path, password string, RSA bool) (KeyPair, error) {
	if RSA {
		privateKey, err := generateRsaKeyPair(2048)
		if err != nil {
			return NewKeyPair(), err
		}
		privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
		if err != nil {
			return NewKeyPair(), err
		}
		publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err != nil {
			return NewKeyPair(), err
		}

		return KeyPair{PrivateKey: privateKeyBytes, PublicKey: publicKeyBytes}, nil

	} else { // Elliptic Curve
		privateKey, err := generateEcKeyPair()
		if err != nil {
			return NewKeyPair(), err
		}

		publicKey := &privateKey.PublicKey

		// what about publicKey.Y.Bytes() ?
		return KeyPair{PrivateKey: privateKey.D.Bytes(), PublicKey: publicKey.X.Bytes()}, nil
	}
}

func generateRsaKeyPair(bits int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func generateEcKeyPair() (*ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func readKeyPairFromFile(path string) (KeyPair, error) {
	var keyPairSerialized KeyPairSerialized

	// Read the JSON file
	data, err := os.ReadFile(path)
	if err != nil {
		return NewKeyPair(), err
	}

	// Unmarshal the JSON data into the serialized struct
	err = json.Unmarshal(data, &keyPairSerialized)
	if err != nil {
		return NewKeyPair(), err
	}

	privateKey, err := hex.DecodeString(keyPairSerialized.PrivateKey)
	if err != nil {
		return NewKeyPair(), err
	}

	publicKey, err := hex.DecodeString(keyPairSerialized.PublicKey)
	if err != nil {
		return NewKeyPair(), err
	}

	return KeyPair{PrivateKey: privateKey, PublicKey: publicKey}, nil
}

func readWalletFromFile(path, pwd string) (Wallet, error) {

	kp, err := readKeyPairFromFile(path)
	if err != nil {
		return NewWallet(), err
	}

	decryptedPrivateKey, err := security.AesEncrypt(kp.PrivateKey, pwd)
	if err != nil {
		return NewWallet(), err
	}

	return Wallet{Key: KeyPair{PublicKey: kp.PublicKey, PrivateKey: decryptedPrivateKey}}, nil
}

func (w Wallet) saveToFile(path, pwd string) error {
	encryptedPrivateKey, err := security.AesEncrypt(w.Key.PrivateKey, pwd)
	if err != nil {
		return err
	}

	publicKeyHex := hex.EncodeToString(w.Key.PublicKey)
	encryptedprivateKeyHex := hex.EncodeToString(encryptedPrivateKey)

	keyPairSerialized := KeyPairSerialized{PrivateKey: encryptedprivateKeyHex, PublicKey: publicKeyHex}

	data, err := json.Marshal(keyPairSerialized)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

/*
func NewKeystoreAccount(dataDir, password string) (database.Address, error) {
	ks := keystore.NewKeyStore(GetKeystoreDirPath(dataDir), keystore.StandardScryptN, keystore.StandardScryptP)
	acc, err := ks.NewAccount(password)
	if err != nil {
		return database.Address{}, err
	}

	return acc.Address, nil
}

func SignTxWithKeystoreAccount(tx database.Tx, acc database.Address, pwd, keystoreDir string) (database.SignedTx, error) {
	ks := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
	ksAccount, err := ks.Find(accounts.Account{Address: acc})
	if err != nil {
		return database.SignedTx{}, err
	}

	ksAccountJson, err := os.ReadFile(ksAccount.URL.Path)
	if err != nil {
		return database.SignedTx{}, err
	}

	key, err := keystore.DecryptKey(ksAccountJson, pwd)
	if err != nil {
		return database.SignedTx{}, err
	}

	signedTx, err := SignTx(tx, key.PrivateKey)
	if err != nil {
		return database.SignedTx{}, err
	}

	return signedTx, nil
}

func SignTx(tx database.Tx, privKey *ecdsa.PrivateKey) (database.SignedTx, error) {
	rawTx, err := tx.Encode()
	if err != nil {
		return database.SignedTx{}, err
	}

	sig, err := Sign(rawTx, privKey)
	if err != nil {
		return database.SignedTx{}, err
	}

	return database.NewSignedTx(tx, sig), nil
}

func Sign(msg []byte, privKey *ecdsa.PrivateKey) (sig []byte, err error) {
	msgHash := sha256.Sum256(msg)

	return crypto.Sign(msgHash[:], privKey)
}

func Verify(msg, sig []byte) (*ecdsa.PublicKey, error) {
	msgHash := sha256.Sum256(msg)

	recoveredPubKey, err := crypto.SigToPub(msgHash[:], sig)
	if err != nil {
		return nil, fmt.Errorf("unable to verify message signature. %s", err.Error())
	}

	return recoveredPubKey, nil
}



func NewRandomKey() (*keystore.Key, error) {
	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	key := &keystore.Key{
		Id:         id,
		Address:    crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		PrivateKey: privateKeyECDSA,
	}

	return key, nil
}

*/
