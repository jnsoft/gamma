package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jnsoft/gamma/src/pkg/common"
	"github.com/jnsoft/gamma/src/pkg/domain/address"
	"github.com/jnsoft/gamma/src/pkg/domain/block"
	"github.com/jnsoft/gamma/src/pkg/domain/genesis"
	"github.com/jnsoft/gamma/src/pkg/domain/state"
)

const (
	DB_FILE       = "block.db"
	BALANCES_FILE = "balances.json"
	NOUNCES_FILE  = "nonces.json"
	GENESIS_FILE  = "genesis.json"
)

func VerifyBlockChain(dataDir string, blocks []block.Block) error {
	jsonBlocks, err := ReadBlocksAsJson(dataDir)
	if err != nil {
		return fmt.Errorf("cannot read blocks from database: %w", err)
	}

	if len(jsonBlocks) == 0 {
		return fmt.Errorf("no blocks in database")
	}

	for i, line := range jsonBlocks {
		var b block.Block
		if err := json.Unmarshal([]byte(line), &b); err != nil {
			return fmt.Errorf("cannot unmarshal block %d: %w", i, err)
		}
		// TODO: link/hash checks, etc.
		_ = b
	}
	return nil
}

func GetStateFromDisk(dataDir string) (*state.State, error) {
	if !common.FileExists(dataDir + DB_FILE) {
		return nil, fmt.Errorf("database file not found")
	}

	statefile, err := os.ReadFile(dataDir + DB_FILE)
	if err != nil {
		return nil, fmt.Errorf("cannot read database file: %w", err)
	}

	state, err := state.NewStateFromJson(string(statefile))
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal database file: %w", err)
	}

	return state, nil
}

func InitDataDirectory(dataDir string) error {
	if err := common.IsWritableDir(dataDir); err != nil {
		return err
	}

	if !common.FileExists(dataDir + GENESIS_FILE) {
		// create genesis file
		if err := os.WriteFile(dataDir+GENESIS_FILE, []byte(genesis.GenesisJson), 0644); err != nil {
			return fmt.Errorf("cannot create genesis file: %w", err)
		}
	}

	if !common.FileExists(dataDir + DB_FILE) {
		// create first block from genesis file and write first block to db
		genesisData, err := os.ReadFile(dataDir + GENESIS_FILE)
		if err != nil {
			return fmt.Errorf("cannot read genesis file: %w", err)
		}

		var g genesis.Genesis
		if err := json.Unmarshal(genesisData, &g); err != nil {
			return fmt.Errorf("cannot unmarshal genesis: %w", err)
		}

		st := state.NewState()

		// g.Balances is map[string]uint, convert keys to address.Address
		for s, bal := range g.Balances {
			var addr address.Address
			copy(addr[:], []byte(s))
			st.Balances[addr] = bal
		}

		sroot := st.ComputeStateRoot()
		var zeroHash [32]byte

		genesisBlock, err := block.NewBlock(zeroHash, sroot, 0, nil, nil)
		if err != nil {
			return fmt.Errorf("cannot create genesis block: %w", err)
		}
		encodedBlock, err := genesisBlock.Encode()
		if err != nil {
			return fmt.Errorf("cannot encode genesis block: %w", err)
		}

		if err := os.WriteFile(dataDir+DB_FILE, append(encodedBlock, '\n'), 0644); err != nil {
			return fmt.Errorf("cannot write database file: %w", err)
		}

		if err := SaveAddressMap(dataDir+BALANCES_FILE, st.Balances); err != nil {
			return fmt.Errorf("cannot save initial balances: %w", err)
		}
		if err := SaveAddressMap(dataDir+NOUNCES_FILE, st.Account2Nonce); err != nil {
			return fmt.Errorf("cannot save initial nonces: %w", err)
		}
	}

	return nil
}

func LoadState(dataDir string) (*state.State, error) {
	if !common.FileExists(dataDir + DB_FILE) {
		return nil, fmt.Errorf("database file not found")
	}

	// Try snapshot files first
	balances, err := GetAddressMap(dataDir + BALANCES_FILE)
	if err != nil {
		// if snapshot missing/corrupt, fall back to empty and rely on replay if you want
		balances = make(map[address.Address]uint)
	}
	nonces, err := GetAddressMap(dataDir + NOUNCES_FILE)
	if err != nil {
		nonces = make(map[address.Address]uint)
	}

	// Determine tip hash & number from the block DB
	f, err := os.Open(dataDir + DB_FILE)
	if err != nil {
		return nil, fmt.Errorf("cannot open block DB file: %w", err)
	}
	defer f.Close()

	var lastBlock block.Block
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		var b block.Block
		if err := json.Unmarshal(line, &b); err != nil {
			return nil, fmt.Errorf("cannot unmarshal block: %w", err)
		}
		lastBlock = b
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning block DB file: %w", err)
	}

	bhash, err := lastBlock.Hash()
	if err != nil {
		return nil, fmt.Errorf("cannot hash last block: %w", err)
	}

	st := state.NewStateWithData(
		balances,
		nonces,
		bhash,
		lastBlock.Header.Number,
	)

	return st, nil
}

// PersistState:
// 1) asks State to produce a new block and update its tip,
// 2) appends that block to DB,
// 3) snapshots balances/nonces.
func PersistState(dataDir string, s *state.State) error {
	blk, err := s.Persist()
	if err != nil {
		return err
	}

	blkBytes, err := blk.Encode()
	if err != nil {
		return fmt.Errorf("cannot encode block: %w", err)
	}

	if err := AppendBlock(dataDir, string(blkBytes)); err != nil {
		return err
	}

	if err := SaveAddressMap(dataDir+BALANCES_FILE, s.Balances); err != nil {
		return fmt.Errorf("cannot save balances: %w", err)
	}
	if err := SaveAddressMap(dataDir+NOUNCES_FILE, s.Account2Nonce); err != nil {
		return fmt.Errorf("cannot save nonces: %w", err)
	}

	return nil
}

func AppendBlock(dataDir, json string) error {
	encoded := []byte(json)

	f, err := os.OpenFile(dataDir+DB_FILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("cannot open block DB file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(encoded); err != nil {
		return fmt.Errorf("cannot write block to DB file: %w", err)
	}
	if _, err := f.Write([]byte("\n")); err != nil {
		return fmt.Errorf("cannot write newline to DB file: %w", err)
	}

	return nil
}

func ReadBlocksAsJson(dataDir string) ([]string, error) {
	var blocks []string

	data, err := os.ReadFile(dataDir + DB_FILE)
	if err != nil {
		return nil, fmt.Errorf("cannot read block DB file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		blocks = append(blocks, line)
	}

	return blocks, nil
}

func SaveAddressMap(path string, balances map[address.Address]uint) error {
	// convert keys to string for JSON
	stringMap := make(map[string]uint, len(balances))
	for addr, bal := range balances {
		stringMap[string(addr[:])] = bal
	}

	data, err := json.Marshal(stringMap)
	if err != nil {
		return fmt.Errorf("cannot marshal balances to JSON: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("cannot write balances to file: %w", err)
	}
	return nil
}

func GetAddressMap(path string) (map[address.Address]uint, error) {
	if !common.FileExists(path) {
		return nil, fmt.Errorf("database file not found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read database file: %w", err)
	}

	// read string-keyed map from JSON
	var stringMap map[string]uint
	if err := json.Unmarshal(data, &stringMap); err != nil {
		return nil, fmt.Errorf("cannot unmarshal database file: %w", err)
	}

	// convert back to address.Address keys
	balances := make(map[address.Address]uint, len(stringMap))
	for s, bal := range stringMap {
		var addr address.Address
		copy(addr[:], []byte(s))
		balances[addr] = bal
	}

	return balances, nil
}
