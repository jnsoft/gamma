package keystore

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestIsUnlocked_Unlock_Lock(t *testing.T) {
	ks, err := NewFileKeystore([]byte("pass"))
	if err != nil {
		t.Fatalf("NewFileKeystore error: %v", err)
	}
	if ks.IsUnlocked() {
		t.Fatalf("expected locked initially")
	}
	if err := ks.Unlock([]byte("pass")); err != nil {
		t.Fatalf("Unlock error: %v", err)
	}
	if !ks.IsUnlocked() {
		t.Fatalf("expected unlocked after Unlock")
	}
	ks.Lock()
	if ks.IsUnlocked() {
		t.Fatalf("expected locked after Lock")
	}
}

func TestFileKeystore_Sign_WithUnlockedKeystore(t *testing.T) {
	// Create keystore
	ks, err := NewFileKeystore([]byte("testpass"))
	if err != nil {
		t.Fatalf("NewFileKeystore error: %v", err)
	}

	// Unlock
	if err := ks.Unlock([]byte("testpass")); err != nil {
		t.Fatalf("Unlock error: %v", err)
	}

	// Build a BIP44-like derivation path
	path := DerivationPath{
		Hardened(44), // purpose
		Hardened(0),  // coin type
		Hardened(0),  // account
		0,            // change
		0,            // index
	}

	// Derive an address to ensure path is usable
	addr, err := ks.DeriveAddress(path)
	if err != nil {
		t.Fatalf("DeriveAddress error: %v", err)
	}
	if addr.String() == "" {
		t.Fatalf("empty address")
	}

	// Message to sign (32 bytes typical hash)
	msg := []byte("0123456789abcdef0123456789abcdef")
	if len(msg) == 0 {
		t.Fatalf("test msg empty")
	}

	// Sign
	sig, err := ks.Sign(path, msg)
	if err != nil {
		t.Fatalf("Sign error: %v", err)
	}
	if len(sig) == 0 {
		t.Fatalf("empty signature")
	}

	t.Logf("addr=%s sig=%s", addr.String(), hex.EncodeToString(sig))
}

func TestFileKeystore_Sign_LockedKeystoreFails(t *testing.T) {
	ks, err := NewFileKeystore([]byte("anotherpass"))
	if err != nil {
		t.Fatalf("NewFileKeystore error: %v", err)
	}

	path := DerivationPath{Hardened(44), Hardened(0), Hardened(0), 0, 1}
	msg := []byte("deadbeefdeadbeefdeadbeefdeadbeef")

	// Attempt to sign without Unlock
	if _, err := ks.Sign(path, msg); err == nil {
		t.Fatalf("expected error signing while locked, got nil")
	}
}

func TestFileKeystore_DeriveAddress_LockedKeystoreFails(t *testing.T) {
	ks, err := NewFileKeystore([]byte("pass"))
	if err != nil {
		t.Fatalf("NewFileKeystore error: %v", err)
	}

	path := DerivationPath{Hardened(44), Hardened(0), Hardened(0), 0, 0}
	if _, err := ks.DeriveAddress(path); err == nil {
		t.Fatalf("expected error deriving while locked, got nil")
	}
}

func TestExportSeed_WhenUnlocked_ReturnsCopy(t *testing.T) {
	ks, err := NewFileKeystore([]byte("pass"))
	if err != nil {
		t.Fatalf("NewFileKeystore error: %v", err)
	}
	// locked -> should fail
	if _, err := ks.ExportSeed(); err == nil {
		t.Fatalf("expected error when exporting locked keystore")
	}
	// unlock
	if err := ks.Unlock([]byte("pass")); err != nil {
		t.Fatalf("Unlock error: %v", err)
	}
	seed, err := ks.ExportSeed()
	if err != nil {
		t.Fatalf("ExportSeed error: %v", err)
	}
	if len(seed) == 0 {
		t.Fatalf("ExportSeed returned empty seed")
	}
	// ensure it's a copy (mutating export must not alter internal seed)
	if len(ks.seed) == 0 {
		t.Fatalf("internal seed unexpectedly empty after export")
	}
	seed[0] ^= 0xFF
	if bytes.Equal(seed, ks.seed) {
		t.Fatalf("exported seed must be a distinct copy")
	}
}

func TestChangePassword_ReencryptsAndKeepsAccess(t *testing.T) {
	ks, err := NewFileKeystore([]byte("oldpass"))
	if err != nil {
		t.Fatalf("NewFileKeystore error: %v", err)
	}
	// Unlock with old password
	if err := ks.Unlock([]byte("oldpass")); err != nil {
		t.Fatalf("Unlock(old) error: %v", err)
	}
	// Change password
	if err := ks.ChangePassword([]byte("oldpass"), []byte("newpass")); err != nil {
		t.Fatalf("ChangePassword error: %v", err)
	}
	// Should be unlocked with updated cached seed
	if !ks.IsUnlocked() {
		t.Fatalf("expected unlocked after ChangePassword")
	}
	// Lock and validate old password fails, new password succeeds
	ks.Lock()
	if ks.IsUnlocked() {
		t.Fatalf("expected locked after Lock")
	}
	if err := ks.Unlock([]byte("oldpass")); err == nil {
		t.Fatalf("expected old password to fail after change")
	}
	if err := ks.Unlock([]byte("newpass")); err != nil {
		t.Fatalf("expected new password to work: %v", err)
	}
}

func TestSaveAndLoad_RestoresEncryptedFields(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "wallet.json")

	ks, err := NewFileKeystore([]byte("pass"))
	if err != nil {
		t.Fatalf("NewFileKeystore error: %v", err)
	}
	if err := ks.Save(file); err != nil {
		t.Fatalf("Save error: %v", err)
	}
	// ensure file exists
	if _, err := os.Stat(file); err != nil {
		t.Fatalf("keystore file not written: %v", err)
	}

	ks2, err := LoadFileKeystore(file)
	if err != nil {
		t.Fatalf("LoadFileKeystore error: %v", err)
	}
	// Basic invariants: encrypted fields present, not unlocked
	if len(ks2.EncSeed) == 0 || len(ks2.Nonce) == 0 || len(ks2.Salt) == 0 {
		t.Fatalf("loaded keystore missing encrypted fields")
	}
	if ks2.IsUnlocked() {
		t.Fatalf("loaded keystore must be locked")
	}
	// Unlock with correct password
	if err := ks2.Unlock([]byte("pass")); err != nil {
		t.Fatalf("Unlock loaded keystore failed: %v", err)
	}
}

func TestAccountPathAndReceivingPath(t *testing.T) {
	// coinType=0, account=0 -> base hardened components
	base := AccountPath(0, 0)
	harden := func(x uint32) uint32 { return x | 0x80000000 }

	if len(base) != 3 {
		t.Fatalf("AccountPath len=%d want 3", len(base))
	}
	if base[0] != harden(44) || base[1] != harden(0) || base[2] != harden(0) {
		t.Fatalf("AccountPath unexpected: %#v", base)
	}

	recv := ReceivingPath(0, 1, 0, 5) // coin=0, account=1, change=0, index=5
	if len(recv) != 5 {
		t.Fatalf("ReceivingPath len=%d want 5", len(recv))
	}
	if recv[0] != harden(44) || recv[1] != harden(0) || recv[2] != harden(1) || recv[3] != 0 || recv[4] != 5 {
		t.Fatalf("ReceivingPath unexpected: %#v", recv)
	}
}
