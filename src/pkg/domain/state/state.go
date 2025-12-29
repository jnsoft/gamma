package state

import (
	"encoding/json"
	"fmt"

	"github.com/jnsoft/gamma/src/pkg/domain/address"
	"github.com/jnsoft/gamma/src/pkg/domain/block"
	"github.com/jnsoft/gamma/src/pkg/domain/signedtx"
)

type State struct {
	Balances        map[address.Address]uint
	Account2Nonce   map[address.Address]uint
	CurrentBlock    block.Block
	TransactionPool []signedtx.SignedTx
	DbFile          string
}

func NewState(dbFile string) *State {
	return &State{
		Balances:        make(map[address.Address]uint),
		Account2Nonce:   make(map[address.Address]uint),
		TransactionPool: []signedtx.SignedTx{},
		DbFile:          dbFile,
	}
}

func NewStateFromJson(data string) (*State, error) {
	state := &State{}
	if err := json.Unmarshal([]byte(data), state); err != nil {
		return nil, fmt.Errorf("cannot unmarshal JSON to state: %w", err)
	}
	return state, nil
}

func (s *State) ToJson() (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", fmt.Errorf("cannot marshal state to JSON: %w", err)
	}
	return string(data), nil
}
