package signedtx

import "github.com/jnsoft/gamma/src/pkg/domain/tx"

type SignedTx struct {
	tx.Tx
	Sig []byte `json:"signature"`
}

func NewSignedTx(tx tx.Tx, sig []byte) SignedTx {
	return SignedTx{tx, sig}
}
