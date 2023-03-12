package database

import (
	"encoding/json"
	"io/ioutil"

	"github.com/ethereum/go-ethereum/common"
)

var genesisJson = `{
	"genesis_time": "2023-03-11T00:00:00.000000000Z",
	"chain_id": "the-gamma-ledger",
	"symbol": "TGL",
	"balances": {
	  "0x0000000000000000000000000000000000000000": 1000000
	},
	"fork_tip_1": 35
  }`

type Genesis struct {
	Balances map[common.Address]uint `json:"balances"`
	Symbol   string                  `json:"symbol"`

	ForkTIP1 uint64 `json:"fork_tip_1"`
}

type old_genesis struct {
	Balances map[Account]uint `json:"balances"`
}

func loadGenesis(path string) (Genesis, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return Genesis{}, err
	}

	var loadedGenesis Genesis
	err = json.Unmarshal(content, &loadedGenesis)
	if err != nil {
		return Genesis{}, err
	}

	return loadedGenesis, nil
}

func writeGenesisToDisk(path string, genesis []byte) error {
	return ioutil.WriteFile(path, genesis, 0644)
}
