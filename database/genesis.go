package database

import (
	"encoding/json"
	"os"
)

type Genesis struct {
	//Time     uint64           `json:"time"`
	Symbol   string           `json:"symbol"`
	Balances map[Address]uint `json:"balances"`
	ForkTIP1 uint64           `json:"fork_tip_1"`
}

// "genesis_time": "2023-03-11T00:00:00.000000000Z",
//	"chain_id": "the-gamma-ledger",

var genesisJson = `{
	"symbol": "TGL",
	"balances": {
		"0x0000000000000000000000000000000000000001": 1000000,
		"0x0000000000000000000000000000000000000002": 1
	 },
	"fork_tip_1": 35
  }`

func loadGenesis(path string) (Genesis, error) {
	content, err := os.ReadFile(path)
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
	return os.WriteFile(path, genesis, 0644)
}
