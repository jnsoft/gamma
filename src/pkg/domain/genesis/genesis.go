package genesis

import (
	"github.com/jnsoft/gamma/src/pkg/common"
	"github.com/jnsoft/gamma/src/pkg/domain/address"
)

type Genesis struct {
	Time     uint64          `json:"time"`
	ChainID  string          `json:"chain_id"`
	Symbol   string          `json:"symbol"`
	Balances map[string]uint `json:"balances"`
}

func NewGenesis(chainID, symbol string, balances map[string]uint) *Genesis {
	return &Genesis{
		Time:     uint64(common.NowUnixUTC()),
		ChainID:  chainID,
		Symbol:   symbol,
		Balances: balances,
	}
}

func NewBalances(addr address.Address, amount uint) map[address.Address]uint {
	return map[address.Address]uint{
		addr: amount,
	}
}

var GenesisJson = `{
	"time": 1767222000,
	"chain_id": "the-gamma-ledger",
	"symbol": "TGL",
	"balances": {
		"3f518628e2cf46531799b8949e874972ebee135c": 1000000
	 }
  }`
