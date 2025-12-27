package wallet

import (
	"github.com/jnsoft/gamma/src/domain/address"
	"github.com/jnsoft/gamma/src/domain/tx"
	"github.com/jnsoft/gamma/src/keystore"
)

type Wallet struct {
	ks      keystore.Keystore
	receive uint32
}

func New(ks keystore.Keystore) *Wallet {
	return &Wallet{ks: ks}
}

func (w *Wallet) NewReceivingAddress() (address.Address, error) {
	path := keystore.DerivationPath{
		keystore.Hardened(44),
		keystore.Hardened(0),
		keystore.Hardened(0),
		0,
		w.receive,
	}
	w.receive++
	return w.ks.DeriveAddress(path)
}

func (w *Wallet) SignTx(tx *tx.Tx, path []uint32) error {
	sig, err := w.ks.Sign(path, tx.Hash())
	if err != nil {
		return err
	}
	tx.Signature = sig
	return nil
}
