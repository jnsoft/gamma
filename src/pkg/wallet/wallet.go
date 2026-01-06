package wallet

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jnsoft/gamma/src/pkg/domain/address"
	"github.com/jnsoft/gamma/src/pkg/domain/signedtx"
	"github.com/jnsoft/gamma/src/pkg/domain/tx"
	"github.com/jnsoft/gamma/src/pkg/keystore"
)

type State struct {
	NextIndex map[uint32]map[uint32]uint32 `json:"nextIndex"`
}

type Wallet struct {
	ks keystore.Keystore
	// nextIndex[account][change] = next receiving index
	nextIndex map[uint32]map[uint32]uint32
}

func New(ks keystore.Keystore) *Wallet {
	return &Wallet{ks: ks, nextIndex: make(map[uint32]map[uint32]uint32)}
}

// SaveState writes nextIndex to disk.
func (w *Wallet) SaveState(path string) error {
	st := State{NextIndex: w.nextIndex}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// LoadState restores nextIndex from disk.
func (w *Wallet) LoadState(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		// start fresh if no state file
		w.nextIndex = make(map[uint32]map[uint32]uint32)
		return nil
	}
	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return err
	}
	if st.NextIndex == nil {
		st.NextIndex = make(map[uint32]map[uint32]uint32)
	}
	w.nextIndex = st.NextIndex
	return nil
}

// NextIndex returns the next index for account/change and increments it.
func (w *Wallet) NextIndex(account, change uint32) uint32 {
	if w.nextIndex[account] == nil {
		w.nextIndex[account] = make(map[uint32]uint32)
	}
	idx := w.nextIndex[account][change]
	w.nextIndex[account][change] = idx + 1
	return idx
}

// PeekIndex returns the current index without incrementing.
func (w *Wallet) PeekIndex(account, change uint32) uint32 {
	if w.nextIndex[account] == nil {
		return 0
	}
	return w.nextIndex[account][change]
}

// SetIndex sets the starting index for an account/change (e.g., after scanning).
func (w *Wallet) SetIndex(account, change, next uint32) {
	if w.nextIndex[account] == nil {
		w.nextIndex[account] = make(map[uint32]uint32)
	}
	w.nextIndex[account][change] = next
}

func (w *Wallet) NewReceivingAddress() (address.Address, error) {
	// default: coinType=0, account=0, change=0 (external)
	coinType, account, change := uint32(0), uint32(0), uint32(0)
	index := w.NextIndex(account, change)

	path := keystore.DerivationPath{
		keystore.Hardened(44),
		keystore.Hardened(coinType),
		keystore.Hardened(account),
		change,
		index,
	}
	return w.ks.DeriveAddress(path)
}

func (w *Wallet) NewReceivingAddressWithPath(path keystore.DerivationPath) (address.Address, error) {
	return w.ks.DeriveAddress(path)
}

func (w *Wallet) SignTx(tx *tx.Tx, path []uint32) (signedtx.SignedTx, error) {
	txhash, err := tx.Hash()
	if err != nil {
		return signedtx.SignedTx{}, err
	}
	sig, err := w.ks.Sign(path, txhash[:])
	if err != nil {
		return signedtx.SignedTx{}, err
	}
	stx := signedtx.NewSignedTx(*tx, sig)
	return stx, nil
}

func (w *Wallet) GetDerivationPath(coinType, account, change, index uint32) keystore.DerivationPath {
	// BIP44-like: m / 44' / coin' / account' / change / index
	return keystore.DerivationPath{
		keystore.Hardened(44), // hardened purpose
		keystore.Hardened(coinType),
		keystore.Hardened(account),
		change,
		index,
	}
}

func (w *Wallet) NewAddressFor(account, change uint32) (address.Address, uint32, error) {
	index := w.NextIndex(account, change)
	path := w.GetDerivationPath(0, account, change, index)
	addr, err := w.ks.DeriveAddress(path)
	if err != nil {
		return address.Address{}, 0, err
	}
	return addr, index, nil
}

func (w *Wallet) String() string {
	return fmt.Sprintf("indexes=%v", w.nextIndex)
}
