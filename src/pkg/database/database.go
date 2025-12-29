package database

import (
	"fmt"
	"os"

	"github.com/jnsoft/gamma/src/pkg/common"
	"github.com/jnsoft/gamma/src/pkg/domain/block"
	"github.com/jnsoft/gamma/src/pkg/domain/genesis"
	"github.com/jnsoft/gamma/src/pkg/domain/state"
)

const (
	DB_FILE      = "block.db"
	GENESIS_FILE = "genesis.json"
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

		genesisBlock := block.NewBlock(nil, 0, nil)
		if err := os.WriteFile(dataDir+DB_FILE, genesisBlock.Encode(), 0644); err != nil {
			return fmt.Errorf("cannot write database file: %w", err)
		}
	}

	return nil
}
