package state

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/jnsoft/gamma/src/pkg/crypto"
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

func (s *State) applyTransaction(tx signedtx.SignedTx) bool {
	if s.Balances[tx.From] < tx.Value {
		return false
	}

	if !tx.IsAuthentic() {
		return false
	}

	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value
	return true
}

func (s *State) ComputeStateRoot() [32]byte {
	// For deterministic hashing, sort addresses first
	addrs := make([]address.Address, 0, len(s.Balances))
	for addr := range s.Balances {
		addrs = append(addrs, addr)
	}

	sort.Slice(addrs, func(i, j int) bool {
		return string(addrs[i][:]) < string(addrs[j][:])
	})

	var data []byte
	for _, addr := range addrs {
		bal := s.Balances[addr]

		data = append(data, addr[:]...)

		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(bal))
		data = append(data, buf...)
	}

	hash := crypto.Sha3_256(data)

	var out [32]byte
	copy(out[:], hash)
	return out
}

func (s *State) ProduceBlock() (block.Block, error) {
	var included []signedtx.SignedTx

	for _, tx := range s.TransactionPool {
		if s.applyTransaction(tx) {
			included = append(included, tx)
		}
	}

	phash, err := s.CurrentBlock.Hash()
	if err != nil {
		return block.Block{}, err
	}

	header := block.BlockHeader{
		ParentHash:       phash,
		StateRoot:        s.ComputeStateRoot(),
		TransactionsRoot: ComputeTransactionsRoot(included),
		BlockNumber:      s.CurrentBlock.Header.Number + 1,
		Timestamp:        uint64(time.Now().Unix()),
	}

	// Empty the transaction pool
	s.TransactionPool = nil

	return block.Block{
		Header: header,
		TXs:    included,
	}
}
