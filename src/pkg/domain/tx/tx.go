package tx

import (
	"encoding/json"
	"time"

	"github.com/jnsoft/gamma/src/pkg/crypto"
	"github.com/jnsoft/gamma/src/pkg/domain/address"
)

type Tx struct {
	From     address.Address `json:"from"`
	To       address.Address `json:"to"`
	Gas      uint            `json:"gas"`
	GasPrice uint            `json:"gasPrice"`
	Value    uint            `json:"value"`
	Nonce    uint            `json:"nonce"`
	Data     string          `json:"data"`
	Time     uint64          `json:"time"`
}

func NewTx(from, to address.Address, gas uint, gasPrice uint, value, nonce uint, data string) Tx {
	return Tx{from, to, gas, gasPrice, value, nonce, data, uint64(time.Now().Unix())}
}

func (t Tx) IsMint() bool {
	return t.Data == "mint"
}

func (t Tx) Cost() uint {
	return t.Value + t.GasCost()
}

func (t Tx) GasCost() uint {
	return t.Gas * t.GasPrice
}

func (t Tx) Hash() (crypto.Hash, error) {
	txJson, err := t.Encode()
	if err != nil {
		return crypto.Hash{}, err
	}

	return crypto.Sha3_256(txJson), nil
}

func (t Tx) Encode() ([]byte, error) {
	return json.Marshal(t)
}

func (t Tx) MarshalJSON() ([]byte, error) {
	type txAlias Tx
	return json.Marshal(txAlias(t))
}
