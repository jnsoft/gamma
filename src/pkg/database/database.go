package database

import (
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
		if err := json.Unmarshal([]byte(genesisData), &g); err != nil {
			return err
		}

		state := state.NewState(DB_FILE)

		for addr, bal := range g.Balances {
			state.Balances[addr] = bal
		}
		sroot := state.ComputeStateRoot()
		var zeroHash [32]byte
		genesisBlock, err := block.NewBlock(zeroHash, sroot, 0, nil, nil)
		if err != nil {
			return fmt.Errorf("cannot create genesis block: %w", err)
		}
		encodeBlock, err := genesisBlock.Encode()
		if err != nil {
			return fmt.Errorf("cannot encode genesis block: %w", err)
		}

		if err := os.WriteFile(dataDir+DB_FILE, encodeBlock, 0644); err != nil {
			return fmt.Errorf("cannot write database file: %w", err)
		}
	}

	return nil
}

func AppendBlock(json string) error {
	encoded := []byte(json)

	f, err := os.OpenFile(DB_FILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
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

func ReadBlocksAsJson() ([]string, error) {
	var blocks []string

	data, err := os.ReadFile(DB_FILE)
	if err != nil {
		return nil, fmt.Errorf("cannot read block DB file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		blocks = append(blocks, line)
	}

	return blocks, nil
}

func SaveAddressMap(path string, balances map[address.Address]uint) error {
	data, err := json.Marshal(balances)
	if err != nil {
		return fmt.Errorf("cannot marshal balances to JSON: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
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

	var balances map[address.Address]uint
	if err := json.Unmarshal(data, &balances); err != nil {
		return nil, fmt.Errorf("cannot unmarshal database file: %w", err)
	}

	return balances, nil
}
