package genesis

import "github.com/jnsoft/gamma/src/pkg/domain/address"

type Genesis struct {
	Time     uint64                   `json:"time"`
	Symbol   string                   `json:"symbol"`
	Balances map[address.Address]uint `json:"balances"`
}

var genesisJson = `{
	"symbol": "TGL",
	"balances": {
		"0x0000000000000000000000000000000000000001": 1000000,
		"0x0000000000000000000000000000000000000002": 1
	 },
	"fork_tip_1": 35
  }`
