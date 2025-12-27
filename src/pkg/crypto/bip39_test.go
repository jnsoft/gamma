package crypto

import (
	"crypto/rand"
	"testing"
)

func TestMnemonicRoundTrip(t *testing.T) {
	entropy := make([]byte, 32)
	for i := range entropy {
		entropy[i] = byte(i)
	}
	mn, err := GetMnemonic(entropy)
	if err != nil {
		t.Fatal(err)
	}
	out, err := GetEntropyFromMnemonic(mn)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != len(entropy) {
		t.Fatalf("len mismatch")
	}
	for i := range out {
		if out[i] != entropy[i] {
			t.Fatalf("mismatch at %d", i)
		}
	}
}

func TestMnemonicRoundTrip_RandomEntropy(t *testing.T) {
	entropy := make([]byte, 32)
	if _, err := rand.Read(entropy); err != nil {
		t.Fatal(err)
	}
	mn, err := GetMnemonic(entropy)
	if err != nil {
		t.Fatal(err)
	}
	out, err := GetEntropyFromMnemonic(mn)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != len(entropy) {
		t.Fatalf("len mismatch: got %d want %d", len(out), len(entropy))
	}
	for i := range out {
		if out[i] != entropy[i] {
			t.Fatalf("mismatch at %d", i)
		}
	}
}
