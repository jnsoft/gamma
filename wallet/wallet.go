package wallet

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jnsoft/gamma/database"
	"github.com/jnsoft/gamma/util/security"
	/*"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"*/)

const keystoreDirName = "keystore"

type KeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
}

type Wallet struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

func GetKeystoreDirPath(dataDir string) string {
	return filepath.Join(dataDir, keystoreDirName)
}

func NewKeyPair() KeyPair {
	return KeyPair{PublicKey: []byte{}, PrivateKey: []byte{}}
}

func NewWallet() Wallet {
	return Wallet{PrivateKey: "", PublicKey: ""}
}

func CreateWallet(path, pwd string) (Wallet, error) {
	keypair, err := createKeyPair(path, pwd)
	if err != nil {
		return NewWallet(), err
	}

	encryptedPrivateKey, err := security.AesEncrypt(keypair.PrivateKey, pwd)
	publicKeyHex := hex.EncodeToString(keypair.PublicKey)
	privateKeyHex := hex.EncodeToString(encryptedPrivateKey)

	w := Wallet{PrivateKey: privateKeyHex, PublicKey: publicKeyHex}

	if err := w.saveToFile(path); err != nil {
		return NewWallet(), err
	}
	return w, nil
}

func GetWallet(path, pwd string) (Wallet, error) {
	keyPair, err := readKeyPairFromFile(path, pwd)
	if err != nil {
		fmt.Println("Error:", err)
		panic(errors.New("read wallet failed"))
	}
	return keyPair
}

func keyPairToWallet(keyPair KeyPair) (Wallet, error) {
	privateKey, err := hex.DecodeString(keyPair.PrivateKey)
	if err != nil {
		return Wallet{}, err
	}
	publicKey, err := hex.DecodeString(keyPair.PublicKey)
	if err != nil {
		return Wallet{}, err
	}
	return Wallet{PrivateKey: privateKey, PublicKey: publicKey}, nil
}

func walletToKeyPair(wallet Wallet) KeyPair {
	return KeyPair{PrivateKey: hex.EncodeToString(wallet.PrivateKey), PublicKey: hex.EncodeToString(wallet.PublicKey)}
}

func (w Wallet) SignTx(tx database.Tx, from database.Address, pwd string) {
	panic("not implemented")
}

func createKeyPair(path, password string) (KeyPair, error) {
	privateKey, err := generateKeyPair(2048)
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
}

func readKeyPairFromFile(path string, password string) (KeyPair, error) {
	var keyPair KeyPair
	data, err := os.ReadFile(path)
	if err != nil {
		return keyPair, err
	}
	err = json.Unmarshal(data, &keyPair)
	if err != nil {
		return keyPair, err
	}
	encryptedPrivateKey, err := hex.DecodeString(keyPair.PrivateKey)
	if err != nil {
		return keyPair, err
	}
	decryptedPrivateKey, err := security.AesDecrypt(encryptedPrivateKey, password)
	if err != nil {
		return keyPair, err
	}
	keyPair.PrivateKey = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: decryptedPrivateKey}))
	return keyPair, nil
}

func generateKeyPair(bits int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func (w Wallet) saveToFile(path string) error {
	data, err := json.Marshal(w)
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
