package signedtx

import (
	"github.com/jnsoft/gamma/src/pkg/crypto"
	"github.com/jnsoft/gamma/src/pkg/domain/address"
	"github.com/jnsoft/gamma/src/pkg/domain/tx"
)

type SignedTx struct {
	tx.Tx
	Sig []byte `json:"signature"`
}

func NewSignedTx(tx tx.Tx, sig []byte) SignedTx {
	return SignedTx{tx, sig}
}

// TODO - Test and fix, needs to recover pubkey from sig (ecp256k1 with a recovery id (v)?)
// other pattern can be include pubkey, verify signature, verify pubkey→address,”
func (t SignedTx) IsAuthentic() (bool, error) {
	txHash, err := t.Tx.Hash()
	if err != nil {
		return false, err
	}

	recoveredPubKey, err := crypto.SigToPub(txHash[:], t.Sig)
	if err != nil {
		return false, err
	}

	recoveredPubKeyBytes := crypto.FromECDSAPub(recoveredPubKey)

	recoveredPubKeyBytesHash := crypto.Sha3_256(recoveredPubKeyBytes[1:])
	recoveredAccount := address.PublicKeyToAddress(recoveredPubKeyBytesHash[12:])

	return recoveredAccount.Hex() == t.From.Hex(), nil

}
