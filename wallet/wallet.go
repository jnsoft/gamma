package wallet

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jnsoft/gamma/database"
	"github.com/jnsoft/gamma/util/security"
)

const keystoreDirName = "keystore"

type Key struct {
	Id         uuid.UUID         // Unique identifier
	Address    database.Address  // Ethereum address
	PrivateKey *ecdsa.PrivateKey // ECDSA private key
}

// Wallet is a keypair with an ecrypted private key
type Wallet struct {
	Key Key
	//PrivateKey string `json:"private_key"`
	//PublicKey  string `json:"public_key"`
}

func GetKeystoreDirPath(dataDir string) string {
	return filepath.Join(dataDir, keystoreDirName)
}

func NewKey() *Key {
	ecdsa, _ := security.GenerateKey()
	id, _ := uuid.NewRandom()
	key := &Key{
		Id:         id,
		Address:    database.BytesToAdress(security.PubKeyToAddress(&ecdsa.PublicKey)),
		PrivateKey: ecdsa,
	}
	return key
}

func CreateWallet(path, pwd string) (Wallet, error) {
	keypair, err := createKeyPair(USE_RSA)
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

func (w Wallet) Address() string {
	address := security.PubKeyToAddress(&w.Key.PrivateKey.PublicKey)
	return hex.EncodeToString(address)
}

func (w Wallet) Hex() string {
	return hex.EncodeToString(w.Key.PublicKey)
}

func (w Wallet) PublicKeyString() string {
	return hex.EncodeToString(w.Key.PublicKey)
}

func (w Wallet) PrivateKeyString() string {
	return hex.EncodeToString(w.Key.PrivateKey)
}

func (w Wallet) SignTx(tx database.Tx, from database.Address, pwd string) {
	panic("not implemented")
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
	return os.WriteFile(GetKeystoreDirPath(path), data, 0644)
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





*/
