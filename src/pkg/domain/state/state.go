package state

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/jnsoft/gamma/src/pkg/crypto"
	"github.com/jnsoft/gamma/src/pkg/domain/address"
	"github.com/jnsoft/gamma/src/pkg/domain/block"
	"github.com/jnsoft/gamma/src/pkg/domain/signedtx"
)

type State struct {
	Balances           map[address.Address]uint
	Account2Nonce      map[address.Address]uint
	CurrentBlockHash   [32]byte
	CurrentBlockNumber uint64
	TransactionPool    []signedtx.SignedTx
}

func NewState() *State {
	return &State{
		Balances:           make(map[address.Address]uint),
		Account2Nonce:      make(map[address.Address]uint),
		CurrentBlockHash:   [32]byte{},
		CurrentBlockNumber: 0,
		TransactionPool:    []signedtx.SignedTx{},
	}
}

func NewStateFromJson(data string) (*State, error) {
	state := &State{}
	if err := json.Unmarshal([]byte(data), state); err != nil {
		return nil, fmt.Errorf("cannot unmarshal JSON to state: %w", err)
	}
	return state, nil
}

func NewStateWithData(
	balances map[address.Address]uint,
	nonces map[address.Address]uint,
	currentBlockHash [32]byte,
	currentBlockNumber uint64,
) *State {
	bals := make(map[address.Address]uint, len(balances))
	for k, v := range balances {
		bals[k] = v
	}
	ncs := make(map[address.Address]uint, len(nonces))
	for k, v := range nonces {
		ncs[k] = v
	}

	return &State{
		Balances:           bals,
		Account2Nonce:      ncs,
		CurrentBlockHash:   currentBlockHash,
		CurrentBlockNumber: currentBlockNumber,
		TransactionPool:    []signedtx.SignedTx{},
	}
}

func (s *State) ToJson() (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", fmt.Errorf("cannot marshal state to JSON: %w", err)
	}
	return string(data), nil
}

func (s *State) AddTransaction(tx signedtx.SignedTx) error {
	if !tx.IsAuthentic() {
		return fmt.Errorf("invalid transaction")
	}
	s.Account2Nonce[tx.From]++
	s.TransactionPool = append(s.TransactionPool, tx)
	return nil
}

func (s *State) Persist() (block.Block, error) {
	// 1. Produce a new block from the current transaction pool and balances
	b, err := s.produceBlock()
	if err != nil {
		return block.Block{}, fmt.Errorf("cannot produce block: %w", err)
	}

	// Update in‑memory chain tip
	bhash, err := b.Hash()
	if err != nil {
		return block.Block{}, fmt.Errorf("cannot hash block: %w", err)
	}
	s.CurrentBlockHash = bhash
	s.CurrentBlockNumber = b.Header.Number

	return b, nil
}

func (s *State) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Current Block: %d\n", s.CurrentBlockNumber))
	b.WriteString("Balances:\n")
	for addr, bal := range s.Balances {
		b.WriteString(fmt.Sprintf("  %s: %d\n", addr, bal))
	}
	b.WriteString("Nonces:\n")
	for addr, nonce := range s.Account2Nonce {
		b.WriteString(fmt.Sprintf("  %s: %d\n", addr, nonce))
	}
	return b.String()
}

// TODO sort transactions before applying
func (s *State) produceBlock() (block.Block, error) {
	var included []signedtx.SignedTx

	for _, tx := range s.TransactionPool {
		if s.applyTransaction(tx) {
			included = append(included, tx)
		}
	}

	newBlock, err := block.NewBlock(s.CurrentBlockHash, computeStateRoot(s.Balances), s.CurrentBlockNumber+1, nil, included)
	if err != nil {
		return block.Block{}, err
	}

	// Empty the transaction pool
	s.TransactionPool = nil

	return newBlock, nil
}

func (s *State) applyTransaction(tx signedtx.SignedTx) bool {
	if !tx.IsAuthentic() {
		return false
	}

	expected := s.Account2Nonce[tx.From] + 1
	if tx.Nonce != expected {
		return false
	}

	if s.Balances[tx.From] < tx.Value {
		return false
	}

	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value
	s.Account2Nonce[tx.From]++
	return true
}

func (s *State) ComputeStateRoot() [32]byte {
	return computeStateRoot(s.Balances)
}

// add Account2Nonce?
func computeStateRoot(balances map[address.Address]uint) [32]byte {
	// For deterministic hashing, sort addresses first
	addrs := make([]address.Address, 0, len(balances))
	for addr := range balances {
		addrs = append(addrs, addr)
	}

	sort.Slice(addrs, func(i, j int) bool {
		return string(addrs[i][:]) < string(addrs[j][:])
	})

	var data []byte
	for _, addr := range addrs {
		bal := balances[addr]

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
