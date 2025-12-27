package crypto

import (
	"crypto/sha256"
	"errors"
)

// ExportMnemonic encodes the unlocked 32-byte seed into a 24-word mnemonic using BIP39-like
// entropy+checksum split. This does NOT derive a BIP39 seed with PBKDF2; it’s a human backup
// of the raw entropy you already use.
func GetMnemonic(entropy []byte) (string, error) {
	if len(entropy) != 32 {
		return "", errors.New("unsupported seed length (want 32 bytes)")
	}

	// checksum = first ENT/32 bits of SHA256(entropy), ENT=256 -> CS=8 bits
	sum := sha256.Sum256(entropy)
	cs := int(sum[0]) // top 8 bits

	// build bitstream ENT+CS = 264 bits
	bits := make([]bool, 0, 256+8)
	for _, b := range entropy {
		for i := 7; i >= 0; i-- {
			bits = append(bits, ((b>>uint(i))&1) == 1)
		}
	}
	for i := 7; i >= 0; i-- {
		bits = append(bits, ((byte(cs)>>uint(i))&1) == 1)
	}

	// split into 24 groups of 11 bits -> indices 0..2047
	indices := make([]int, 24)
	for i := 0; i < 24; i++ {
		val := 0
		for j := 0; j < 11; j++ {
			val <<= 1
			if bits[i*11+j] {
				val |= 1
			}
		}
		indices[i] = val
	}

	// map to words
	words := make([]byte, 0, 24*10)
	for i, idx := range indices {
		if idx < 0 || idx >= len(BIP39Wordlist) {
			return "", errors.New("word index out of range")
		}
		word := BIP39Wordlist[idx]
		words = append(words, word...)
		if i != len(indices)-1 {
			words = append(words, ' ')
		}
	}
	return string(words), nil
}

// NewFileKeystoreFromMnemonic decodes 24 words back to raw entropy and stores it;
// does NOT run PBKDF2-HMAC-SHA512. It restores the same 32-byte seed you back up.
func NewFileKeystoreFromMnemonic(mnemonic string) ([]byte, error) {
	if mnemonic == "" {
		return nil, errors.New("empty mnemonic")
	}
	parts := splitWords(mnemonic)
	if len(parts) != 24 {
		return nil, errors.New("expected 24 words for 256-bit entropy")
	}

	// build 264-bit stream from indices
	bits := make([]bool, 0, 264)
	for _, w := range parts {
		idx := findWordIndex(w)
		if idx < 0 {
			return nil, errors.New("unknown word: " + w)
		}
		// 11 bits
		for i := 10; i >= 0; i-- {
			bits = append(bits, ((idx>>uint(i))&1) == 1)
		}
	}
	if len(bits) != 264 {
		return nil, errors.New("bitstream size mismatch")
	}

	// extract entropy (first 256 bits)
	entropy := make([]byte, 32)
	for i := 0; i < 32; i++ {
		var b byte
		for j := 0; j < 8; j++ {
			b <<= 1
			if bits[i*8+j] {
				b |= 1
			}
		}
		entropy[i] = b
	}
	// verify checksum
	sum := sha256.Sum256(entropy)
	cs := byte(0)
	for i := 0; i < 8; i++ {
		cs <<= 1
		if bits[256+i] {
			cs |= 1
		}
	}
	if cs != sum[0]>>0 { // compare first 8 bits
		return nil, errors.New("checksum mismatch")
	}

	return entropy, nil
}

func splitWords(s string) []string {
	out := []string{}
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ' ' || s[i] == '\n' || s[i] == '\t' {
			if start < i {
				out = append(out, s[start:i])
			}
			start = i + 1
		}
	}
	return out
}

func findWordIndex(word string) int {
	// linear search; for performance, build a map[string]int at init
	for i, w := range BIP39Wordlist {
		if w == word {
			return i
		}
	}
	return -1
}
