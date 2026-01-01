package state

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/jnsoft/gamma/src/pkg/crypto"
	"github.com/jnsoft/gamma/src/pkg/database"
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

func NewState(dbFile string) *State {
	balances, err := database.GetBalances()
	if err != nil {
		balances = make(map[address.Address]uint)
	}
	nonces, err := database.GetNonces()
	if err != nil {
		nonces = make(map[address.Address]uint)
	}
	return &State{
		Balances:        balances,
		Account2Nonce:   nonces,
		TransactionPool: []signedtx.SignedTx{},
	}
}

func NewStateFromJson(data string) (*State, error) {
	state := &State{}
	if err := json.Unmarshal([]byte(data), state); err != nil {
		return nil, fmt.Errorf("cannot unmarshal JSON to state: %w", err)
	}
	return state, nil
}

func verifyBlockChain([]block.Block) error {
	jsonBlocks, err := database.ReadBlocksAsJson()
	if err != nil {
		return fmt.Errorf("cannot read blocks from database: %w", err)
	}

	return nil
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

func (s *State) Persist() error {
	// 1. Produce a new block from the current transaction pool and balances
	b, err := s.produceBlock()
	if err != nil {
		return fmt.Errorf("cannot produce block: %w", err)
	}
	bhash, err := b.Hash()
	if err != nil {
		return fmt.Errorf("cannot hash block: %w", err)
	}
	s.CurrentBlockHash = bhash
	s.CurrentBlockNumber++

	// 2. Encode block to JSON and append to block DB file
	blockBytes, err := b.Encode()
	if err != nil {
		return fmt.Errorf("cannot encode block: %w", err)
	}
	if err := database.AppendBlock(string(blockBytes)); err != nil {
		return fmt.Errorf("cannot append block to file: %w", err)
	}

	// 4. Save current Balances
	if err := database.SaveAddressMap(database.BALANCES_FILE, s.Balances); err != nil {
		return fmt.Errorf("cannot save balances: %w", err)
	}

	if err := database.SaveAddressMap(database.NOUNCES_FILE, s.Account2Nonce); err != nil {
		return fmt.Errorf("cannot save nonces: %w", err)
	}

	return nil
}

func (s *State) produceBlock() (block.Block, error) {
	var included []signedtx.SignedTx

	for _, tx := range s.TransactionPool {
		if s.applyTransaction(tx) {
			included = append(included, tx)
		}
	}

	newBlock, err := block.NewBlock(s.CurrentBlockHash, s.computeStateRoot(), s.CurrentBlockNumber+1, nil, included)
	if err != nil {
		return block.Block{}, err
	}

	// Empty the transaction pool
	s.TransactionPool = nil

	return newBlock, nil
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

// add Account2Nonce?
func (s *State) computeStateRoot() [32]byte {
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
