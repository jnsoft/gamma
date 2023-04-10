/*
• Adding new transactions to Mempool
• Validating transactions against the current State (sufficient sender balance)
• Changing the state
• Persisting transactions to disk
• Calculating accounts balances by replaying all transactions since Genesis in a sequence
*/
package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
)

const TxGas = 21
const TxGasPriceDefault = 1
const TxFee = uint(50)

type State struct {
	Balances      map[Address]uint
	Account2Nonce map[Address]uint

	txMempool []SimpleTx // Only for SimpleTx

	dbFile *os.File

	latestBlock     SimpleBlock
	latestBlockHash Hash
	hasGenesisBlock bool

	miningDifficulty uint

	forkTIP1 uint64

	HashCache   map[string]int64
	HeightCache map[uint64]int64
}

func NewStateFromDisk(dataDir string, miningDifficulty uint) (*State, error) {
	err := InitDataDirIfNotExists(dataDir, []byte(genesisJson))
	if err != nil {
		return nil, err
	}

	gen, err := loadGenesis(getGenesisJsonFilePath(dataDir))
	if err != nil {
		return nil, err
	}

	balances := make(map[Address]uint)
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	account2nonce := make(map[Address]uint)

	dbFilepath := getBlocksDbFilePath(dataDir)
	f, err := os.OpenFile(dbFilepath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)

	state := &State{balances, account2nonce, nil, f, SimpleBlock{}, Hash{}, false, miningDifficulty, gen.ForkTIP1, map[string]int64{}, map[uint64]int64{}}

	// set file position
	filePos := int64(0)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		blockFsJson := scanner.Bytes()

		if len(blockFsJson) == 0 {
			break
		}

		var blockFs SimpleBlockFS
		err = json.Unmarshal(blockFsJson, &blockFs)
		if err != nil {
			return nil, err
		}

		//err = applyBlock(blockFs.Value, state)
		err = applySimpleBlock(blockFs.Value, state)

		if err != nil {
			return nil, err
		}

		// set search caches
		state.HashCache[blockFs.Key.Hex()] = filePos
		state.HeightCache[blockFs.Value.Header.Number] = filePos
		filePos += int64(len(blockFsJson)) + 1

		state.latestBlock = blockFs.Value
		state.latestBlockHash = blockFs.Key
		state.hasGenesisBlock = true
	}

	return state, nil
}

/*
	func (s *State) AddBlocks(blocks []Block) error {
		for _, b := range blocks {
			_, err := s.AddBlock(b)
			if err != nil {
				return err
			}
		}

		return nil
	}

	func (s *State) AddBlock(b Block) (Hash, error) {
		pendingState := s.Copy()

		err := applyBlock(b, &pendingState)
		if err != nil {
			return Hash{}, err
		}

		blockHash, err := b.Hash()
		if err != nil {
			return Hash{}, err
		}

		blockFs := BlockFS{blockHash, b}

		blockFsJson, err := json.Marshal(blockFs)
		if err != nil {
			return Hash{}, err
		}

		fmt.Printf("\nPersisting new Block to disk:\n")
		fmt.Printf("\t%s\n", blockFsJson)

		// get file pos for cache
		fs, _ := s.dbFile.Stat()
		filePos := fs.Size() + 1

		_, err = s.dbFile.Write(append(blockFsJson, '\n'))
		if err != nil {
			return Hash{}, err
		}

		// set search caches
		s.HashCache[blockFs.Key.Hex()] = filePos
		s.HeightCache[blockFs.Value.Header.Number] = filePos

		s.Balances = pendingState.Balances
		s.Account2Nonce = pendingState.Account2Nonce
		s.latestBlockHash = blockHash
		s.latestBlock = b
		s.hasGenesisBlock = true
		s.miningDifficulty = pendingState.miningDifficulty

		return blockHash, nil
	}
*/
func (s *State) AddSimpleBlock(b SimpleBlock) error {
	for _, tx := range b.TXs {
		if err := s.AddSimpleTx(tx); err != nil {
			return err
		}
	}

	return nil
}

func (s *State) AddSimpleTx(tx SimpleTx) error {
	if err := s.applySimple(tx); err != nil {
		return err
	}

	s.txMempool = append(s.txMempool, tx)

	return nil
}

func (s *State) applySimple(tx SimpleTx) error {
	if tx.IsMint() {
		s.Balances[tx.To] += tx.Value
		return nil
	}

	if s.Balances[tx.From] < tx.Value {
		return fmt.Errorf("insufficient balance")
	}

	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value

	return nil
}

func (s *State) Persist() (Hash, error) {
	block := NewSimpleBlock(s.LatestBlockHash(), s.NextBlockNumber(), ToAddress(A0), s.txMempool)
	blockHash, err := block.Hash()
	if err != nil {
		return Hash{}, err
	}

	blockFs := SimpleBlockFS{blockHash, block}

	blockFsJson, err := json.Marshal(blockFs)
	if err != nil {
		return Hash{}, err
	}

	fmt.Printf("Persisting new Block to disk:\n")
	fmt.Printf("\t%s\n", blockFsJson)

	if _, err = s.dbFile.Write(append(blockFsJson, '\n')); err != nil {
		return Hash{}, err
	}
	s.latestBlockHash = blockHash

	s.txMempool = []SimpleTx{}

	return s.latestBlockHash, nil
}

func (s *State) NextBlockNumber() uint64 {
	if !s.hasGenesisBlock {
		return uint64(0)
	}

	return s.LatestBlock().Header.Number + 1
}

func (s *State) LatestBlock() SimpleBlock {
	return s.latestBlock
}

func (s *State) LatestBlockHash() Hash {
	return s.latestBlockHash
}

func (s *State) GetNextAccountNonce(account Address) uint {
	return s.Account2Nonce[account] + 1
}

func (s *State) ChangeMiningDifficulty(newDifficulty uint) {
	s.miningDifficulty = newDifficulty
}

func (s *State) IsTIP1Fork() bool {
	return s.NextBlockNumber() >= s.forkTIP1
}

func (s *State) Copy() State {
	c := State{}
	c.hasGenesisBlock = s.hasGenesisBlock
	c.latestBlock = s.latestBlock
	c.latestBlockHash = s.latestBlockHash
	c.Balances = make(map[Address]uint)
	c.Account2Nonce = make(map[Address]uint)
	c.miningDifficulty = s.miningDifficulty
	c.forkTIP1 = s.forkTIP1

	for acc, balance := range s.Balances {
		c.Balances[acc] = balance
	}

	for acc, nonce := range s.Account2Nonce {
		c.Account2Nonce[acc] = nonce
	}

	return c
}

func (s *State) Close() error {
	return s.dbFile.Close()
}

// applyBlock verifies if block can be added to the blockchain.
//
// Block metadata are verified as well as transactions within (sufficient balances, etc).
func applyBlock(b Block, s *State) error {
	nextExpectedBlockNumber := s.latestBlock.Header.Number + 1

	if s.hasGenesisBlock && b.Header.Number != nextExpectedBlockNumber {
		return fmt.Errorf("next expected block must be '%d' not '%d'", nextExpectedBlockNumber, b.Header.Number)
	}

	if s.hasGenesisBlock && s.latestBlock.Header.Number > 0 && !reflect.DeepEqual(b.Header.Parent, s.latestBlockHash) {
		return fmt.Errorf("next block parent hash must be '%x' not '%x'", s.latestBlockHash, b.Header.Parent)
	}

	hash, err := b.Hash()
	if err != nil {
		return err
	}

	if !IsBlockHashValid(hash, s.miningDifficulty) {
		return fmt.Errorf("invalid block hash %x", hash)
	}

	err = applyTXs(b.TXs, s)
	if err != nil {
		return err
	}

	s.Balances[b.Header.Miner] += BlockReward
	if s.IsTIP1Fork() {
		s.Balances[b.Header.Miner] += b.GasReward()
	} else {
		s.Balances[b.Header.Miner] += uint(len(b.TXs)) * TxFee
	}

	return nil
}

func applySimpleBlock(b SimpleBlock, s *State) error {
	nextExpectedBlockNumber := s.NextBlockNumber()

	if s.hasGenesisBlock && b.Header.Number != nextExpectedBlockNumber {
		return fmt.Errorf("next expected block must be '%d' not '%d'", nextExpectedBlockNumber, b.Header.Number)
	}

	if s.hasGenesisBlock && s.latestBlock.Header.Number > 0 && !reflect.DeepEqual(b.Header.Parent, s.latestBlockHash) {
		return fmt.Errorf("next block parent hash must be '%x' not '%x'", s.latestBlockHash, b.Header.Parent)
	}

	hash, err := b.Hash()
	if err != nil {
		return err
	}

	if !IsBlockHashValid(hash, s.miningDifficulty) {
		return fmt.Errorf("invalid block hash %x", hash)
	}

	err = applySimpleTXs(b.TXs, s)
	if err != nil {
		return err
	}

	s.Balances[b.Header.Miner] += BlockReward
	if s.IsTIP1Fork() {
		s.Balances[b.Header.Miner] += b.GasReward()
	} else {
		s.Balances[b.Header.Miner] += uint(len(b.TXs)) * TxFee
	}

	return nil
}

func applyTXs(txs []SignedTx, s *State) error {
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].Time < txs[j].Time
	})

	for _, tx := range txs {
		err := ApplyTx(tx, s)
		if err != nil {
			return err
		}
	}

	return nil
}

func applySimpleTXs(txs []SimpleTx, s *State) error {
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].Time < txs[j].Time
	})

	for _, tx := range txs {
		err := ApplySimpleTx(tx, s)
		if err != nil {
			return err
		}
	}

	return nil
}

func ApplyTx(tx SignedTx, s *State) error {
	err := ValidateTx(tx, s)
	if err != nil {
		return err
	}

	s.Balances[tx.From] -= tx.Cost(s.IsTIP1Fork())
	s.Balances[tx.To] += tx.Value

	s.Account2Nonce[tx.From] = tx.Nonce

	return nil
}

func ApplySimpleTx(tx SimpleTx, s *State) error {
	err := ValidateSimpleTx(tx, s)
	if err != nil {
		return err
	}

	s.Balances[tx.From] -= tx.Cost(s.IsTIP1Fork())
	s.Balances[tx.To] += tx.Value

	s.Account2Nonce[tx.From] = tx.Nonce

	return nil
}

func ValidateTx(tx SignedTx, s *State) error {
	ok, err := tx.IsAuthentic()
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("wrong TX. Sender '%s' is forged", tx.From.String())
	}

	expectedNonce := s.GetNextAccountNonce(tx.From)
	if tx.Nonce != expectedNonce {
		return fmt.Errorf("wrong TX. Sender '%s' next nonce must be '%d', not '%d'", tx.From.String(), expectedNonce, tx.Nonce)
	}

	if s.IsTIP1Fork() {
		// For now we only have one type, transfer TXs, so all TXs must pay 21 gas like on Ethereum (21 000)
		if tx.Gas != TxGas {
			return fmt.Errorf("insufficient TX gas %v. required: %v", tx.Gas, TxGas)
		}

		if tx.GasPrice < TxGasPriceDefault {
			return fmt.Errorf("insufficient TX gasPrice %v. required at least: %v", tx.GasPrice, TxGasPriceDefault)
		}

	} else {
		// Prior to TIP1, a signed TX must NOT populate the Gas fields to prevent consensus from crashing
		// It's not enough to add this validation to http_routes.go because a TX could come from another node
		// that could modify its software and broadcast such a TX, it must be validated here too.
		if tx.Gas != 0 || tx.GasPrice != 0 {
			return fmt.Errorf("invalid TX. `Gas` and `GasPrice` can't be populated before TIP1 fork is active")
		}
	}

	if tx.Cost(s.IsTIP1Fork()) > s.Balances[tx.From] {
		return fmt.Errorf("wrong TX. Sender '%s' balance is %d TBB. Tx cost is %d TBB", tx.From.String(), s.Balances[tx.From], tx.Cost(s.IsTIP1Fork()))
	}

	return nil
}

func ValidateSimpleTx(tx SimpleTx, s *State) error {
	ok, err := tx.IsAuthentic()
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("wrong TX. Sender '%s' is forged", tx.From.String())
	}

	expectedNonce := s.GetNextAccountNonce(tx.From)
	if tx.Nonce != expectedNonce {
		return fmt.Errorf("wrong TX. Sender '%s' next nonce must be '%d', not '%d'", tx.From.String(), expectedNonce, tx.Nonce)
	}

	if tx.Cost(s.IsTIP1Fork()) > s.Balances[tx.From] {
		return fmt.Errorf("wrong TX. Sender '%s' balance is %d TBB. Tx cost is %d TBB", tx.From.String(), s.Balances[tx.From], tx.Cost(s.IsTIP1Fork()))
	}

	return nil
}
