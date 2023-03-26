package database

import (
	"encoding/json"
	"io/ioutil"
)

var genesisJson = `{
	"genesis_time": "2023-03-11T00:00:00.000000000Z",
	"chain_id": "the-gamma-ledger",
	"symbol": "TGL",
	"balances": {
	  "0x1": 1000000,
	  "0x2": 1
	},
	"fork_tip_1": 35
  }`

type Genesis struct {
	Balances map[Address]uint `json:"balances"`
	Symbol   string           `json:"symbol"`

	ForkTIP1 uint64 `json:"fork_tip_1"`
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
