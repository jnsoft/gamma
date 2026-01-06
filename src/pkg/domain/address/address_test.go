package address

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestToAddress(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		wantError bool
	}{
		{
			name:      "Valid 20 bytes",
			input:     make([]byte, 20),
			wantError: false,
		},
		{
			name:      "Invalid length - empty",
			input:     []byte{},
			wantError: true,
		},
		{
			name:      "Invalid length - too short (19 bytes)",
			input:     make([]byte, 19),
			wantError: true,
		},
		{
			name:      "Invalid length - too long (21 bytes)",
			input:     make([]byte, 21),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToAddress(tt.input)
			if tt.wantError {
				if err == nil {
					t.Errorf("ToAddress() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ToAddress() unexpected error: %v", err)
				}
				if !bytes.Equal(got[:], tt.input) {
					t.Errorf("ToAddress() = %x, want %x", got, tt.input)
				}
			}
		})
	}
}

func TestHexToAddress(t *testing.T) {
	validHex := "0102030405060708090a0b0c0d0e0f1011121314"

	t.Run("Valid hex string", func(t *testing.T) {
		addr, err := HexToAddress(validHex)
		if err != nil {
			t.Fatalf("HexToAddress failed: %v", err)
		}
		if addr.Hex() != validHex {
			t.Errorf("Hex() got %s, want %s", addr.Hex(), validHex)
		}
		// String() usually returns the same as Hex()
		if addr.String() != validHex {
			t.Errorf("String() got %s, want %s", addr.String(), validHex)
		}
	})

	t.Run("Invalid hex characters", func(t *testing.T) {
		_, err := HexToAddress("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
		if err == nil {
			t.Error("expected error for invalid hex chars, got nil")
		}
	})

	t.Run("Valid hex length but incorrect address length (short)", func(t *testing.T) {
		// "001122" decodes to 3 bytes, ToAddress expects 20
		_, err := HexToAddress("001122")
		if err == nil {
			t.Error("expected error for short byte sequence, got nil")
		}
	})

	t.Run("Odd length string", func(t *testing.T) {
		_, err := HexToAddress("123")
		if err == nil {
			t.Error("expected error for odd length string, got nil")
		}
	})
}

func TestPublicKeyToAddress(t *testing.T) {
	// Simulate an uncompressed public key: 1 byte prefix (0x04) + 64 bytes data
	pubKey := make([]byte, 65)
	pubKey[0] = 0x04
	for i := 1; i < 65; i++ {
		pubKey[i] = byte(i) // fill with predictable data
	}

	addr := PublicKeyToAddress(pubKey)

	// Since PublicKeyToAddress uses crypto.Sha3_256 internally,
	// we just verify basic properties here (e.g. deterministic output).

	// 1. determinism
	addr2 := PublicKeyToAddress(pubKey)
	if addr != addr2 {
		t.Error("PublicKeyToAddress is not deterministic")
	}

	// 2. non-empty check (it's extremely unlikely a real hash is all zeros)
	allZeros := true
	for _, b := range addr {
		if b != 0 {
			allZeros = false
			break
		}
	}
	if allZeros {
		t.Error("PublicKeyToAddress returned all zeros, seems incorrect")
	}
}

func TestAddress_Formatting(t *testing.T) {
	// Create address [0, 1, 2, ... 19]
	var raw [20]byte
	for i := range raw {
		raw[i] = byte(i)
	}
	addr, _ := ToAddress(raw[:])

	expected := hex.EncodeToString(raw[:])

	if got := addr.Hex(); got != expected {
		t.Errorf("Hex() = %s, want %s", got, expected)
	}

	if got := addr.String(); got != expected {
		t.Errorf("String() = %s, want %s", got, expected)
	}
}
