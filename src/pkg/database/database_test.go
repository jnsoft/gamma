package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jnsoft/gamma/src/pkg/domain/address"
)

func newTempDir(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "gamma-db-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	// make sure we always have a trailing slash like the rest of the code expects
	return dir + string(os.PathSeparator)
}

func TestInitDataDirectory_CreatesGenesisAndDb(t *testing.T) {
	dataDir := newTempDir(t)
	defer os.RemoveAll(dataDir)

	if err := InitDataDirectory(dataDir); err != nil {
		t.Fatalf("InitDataDirectory failed: %v", err)
	}

	// expect all core files to exist
	for _, f := range []string{GENESIS_FILE, DB_FILE, BALANCES_FILE, NOUNCES_FILE} {
		path := filepath.Join(dataDir, f)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to be created, got error: %v", path, err)
		}
	}

	// db file should not be empty (must contain genesis block)
	dbPath := filepath.Join(dataDir, DB_FILE)
	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("stat on %s failed: %v", dbPath, err)
	}
	if info.Size() == 0 {
		t.Fatalf("expected non‑empty %s (genesis block), size is 0", dbPath)
	}
}

func TestLoadState_InitialGenesisState(t *testing.T) {
	dataDir := newTempDir(t)
	defer os.RemoveAll(dataDir)

	if err := InitDataDirectory(dataDir); err != nil {
		t.Fatalf("InitDataDirectory failed: %v", err)
	}

	st, err := LoadState(dataDir)
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if st == nil {
		t.Fatalf("LoadState returned nil state")
	}

	// after InitDataDirectory we should have at least the genesis block
	if st.CurrentBlockNumber != 0 {
		t.Fatalf("expected CurrentBlockNumber 0, got %d", st.CurrentBlockNumber)
	}
}

func TestPersistState_AppendsBlockAndUpdatesSnapshots(t *testing.T) {
	dataDir := newTempDir(t)
	defer os.RemoveAll(dataDir)

	if err := InitDataDirectory(dataDir); err != nil {
		t.Fatalf("InitDataDirectory failed: %v", err)
	}

	st, err := LoadState(dataDir)
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	// give some balance to a fake address so we can create a valid tx if needed later
	var addr address.Address
	copy(addr[:], []byte("addr-1"))
	st.Balances[addr] = 100

	// persist without any txs first, should just create an empty block
	if err := PersistState(dataDir, st); err != nil {
		t.Fatalf("PersistState failed: %v", err)
	}

	// verify DB file has at least 2 lines (genesis + new block)
	dbPath := filepath.Join(dataDir, DB_FILE)
	data, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("cannot read db file: %v", err)
	}

	blocks, err := ReadBlocksAsJson(dataDir)
	if err != nil {
		t.Fatalf("ReadBlocksAsJson failed: %v", err)
	}
	if len(blocks) < 2 {
		t.Fatalf("expected at least 2 blocks (genesis + 1), got %d", len(blocks))
	}
	if len(data) == 0 {
		t.Fatalf("db file should not be empty")
	}

	// verify balances / nonces snapshots were updated and are readable
	balances, err := GetAddressMap(dataDir + BALANCES_FILE)
	if err != nil {
		t.Fatalf("GetAddressMap(balances) failed: %v", err)
	}
	if got := balances[addr]; got != 100 {
		t.Fatalf("expected balance %d for addr, got %d", 100, got)
	}

	nonces, err := GetAddressMap(dataDir + NOUNCES_FILE)
	if err != nil {
		t.Fatalf("GetAddressMap(nonces) failed: %v", err)
	}

	// We haven't sent any tx from addr, so nonce should just be zero.
	if got := nonces[addr]; got != 0 {
		t.Fatalf("expected nonce 0 for addr, got %d", got)
	}
}

func TestSaveAndGetAddressMap_RoundTrip(t *testing.T) {
	dataDir := newTempDir(t)
	defer os.RemoveAll(dataDir)

	path := filepath.Join(dataDir, "test-balances.json")

	orig := make(map[address.Address]uint)
	var a1, a2 address.Address
	copy(a1[:], []byte("a1"))
	copy(a2[:], []byte("a2"))
	orig[a1] = 10
	orig[a2] = 20

	if err := SaveAddressMap(path, orig); err != nil {
		t.Fatalf("SaveAddressMap failed: %v", err)
	}

	got, err := GetAddressMap(path)
	if err != nil {
		t.Fatalf("GetAddressMap failed: %v", err)
	}

	if len(got) != len(orig) {
		t.Fatalf("expected %d entries, got %d", len(orig), len(got))
	}

	if got[a1] != 10 || got[a2] != 20 {
		t.Fatalf("unexpected balances after round‑trip: %+v", got)
	}
}

// optional: sanity test for PersistState & LoadState integration
func TestPersistAndReloadState_Integration(t *testing.T) {
	dataDir := newTempDir(t)
	defer os.RemoveAll(dataDir)

	if err := InitDataDirectory(dataDir); err != nil {
		t.Fatalf("InitDataDirectory failed: %v", err)
	}

	st, err := LoadState(dataDir)
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	// mutate in‑memory state
	var addr address.Address
	copy(addr[:], []byte("integration-addr"))
	st.Balances[addr] = 42

	if err := PersistState(dataDir, st); err != nil {
		t.Fatalf("PersistState failed: %v", err)
	}

	// reload and check that snapshot was picked up
	st2, err := LoadState(dataDir)
	if err != nil {
		t.Fatalf("LoadState (2) failed: %v", err)
	}

	if got := st2.Balances[addr]; got != 42 {
		t.Fatalf("expected reloaded balance 42, got %d", got)
	}

	// block number should have advanced by 1
	if st2.CurrentBlockNumber != st.CurrentBlockNumber {
		t.Fatalf("expected CurrentBlockNumber %d, got %d", st.CurrentBlockNumber, st2.CurrentBlockNumber)
	}
}
