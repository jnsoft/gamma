package wallet

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/jnsoft/gamma/src/pkg/domain/address"
	"github.com/jnsoft/gamma/src/pkg/domain/signedtx"
	"github.com/jnsoft/gamma/src/pkg/domain/tx"
	"github.com/jnsoft/gamma/src/pkg/keystore"
)

type ksStub struct {
	deriveCalls [][]uint32
	signCalls   [][]uint32
	signMsgs    [][]byte
	// configurable outputs
	deriveAddr address.Address
	signSig    []byte
}

func (s *ksStub) DeriveAddress(path keystore.DerivationPath) (address.Address, error) {
	// record
	cp := append([]uint32(nil), path...)
	s.deriveCalls = append(s.deriveCalls, cp)
	// default deterministic addr if not set: encode index in last byte
	if s.deriveAddr == (address.Address{}) {
		var a address.Address
		if len(path) > 0 {
			a[len(a)-1] = byte(path[len(path)-1] & 0xFF)
		}
		return a, nil
	}
	return s.deriveAddr, nil
}

func (s *ksStub) Sign(path keystore.DerivationPath, msg []byte) ([]byte, error) {
	// record
	s.signCalls = append(s.signCalls, append([]uint32(nil), path...))
	s.signMsgs = append(s.signMsgs, append([]byte(nil), msg...))
	if len(s.signSig) == 0 {
		return []byte{0xDE, 0xAD, 0xBE, 0xEF}, nil
	}
	return s.signSig, nil
}

func mustAddress(b byte) address.Address {
	var a address.Address
	for i := range a {
		a[i] = b
	}
	return a
}

// helper to compare uint32 slices
func u32sToBytes(v []uint32) []byte {
	b := make([]byte, 0, 4*len(v))
	for _, x := range v {
		b = append(b, byte(x>>24), byte(x>>16), byte(x>>8), byte(x))
	}
	return b
}

func TestNewReceivingAddress_IncrementsIndexAndPath(t *testing.T) {
	fk := &ksStub{deriveAddr: mustAddress(0xAA)}
	w := New(fk)

	// First address
	addr1, err := w.NewReceivingAddress()
	if err != nil {
		t.Fatalf("NewReceivingAddress 1 error: %v", err)
	}
	if addr1 != fk.deriveAddr {
		t.Fatalf("unexpected addr1: %v", addr1)
	}
	if len(fk.deriveCalls) != 1 {
		t.Fatalf("expected 1 derive call, got %d", len(fk.deriveCalls))
	}
	path1 := fk.deriveCalls[0]
	// Expected BIP44-like: 44' / 0' / 0' / 0 / index(0)
	if len(path1) != 5 {
		t.Fatalf("path1 len = %d, want 5", len(path1))
	}
	harden := func(x uint32) uint32 { return x | 0x80000000 }
	if path1[0] != harden(44) || path1[1] != harden(0) || path1[2] != harden(0) || path1[3] != 0 || path1[4] != 0 {
		t.Fatalf("unexpected path1: %#v", path1)
	}

	// Second address
	addr2, err := w.NewReceivingAddress()
	if err != nil {
		t.Fatalf("NewReceivingAddress 2 error: %v", err)
	}
	if addr2 != fk.deriveAddr {
		t.Fatalf("unexpected addr2: %v", addr2)
	}
	if len(fk.deriveCalls) != 2 {
		t.Fatalf("expected 2 derive calls, got %d", len(fk.deriveCalls))
	}
	path2 := fk.deriveCalls[1]
	if path2[4] != 1 {
		t.Fatalf("receive index did not increment, got %d", path2[4])
	}
}

func TestNewReceivingAddressWithPath_ForwardsPath(t *testing.T) {
	fk := &ksStub{deriveAddr: mustAddress(0xBB)}
	w := New(fk)

	path := keystore.DerivationPath{
		keystore.Hardened(44),
		keystore.Hardened(1),
		keystore.Hardened(2),
		3,
		4,
	}

	addr, err := w.NewReceivingAddressWithPath(path)
	if err != nil {
		t.Fatalf("NewReceivingAddressWithPath error: %v", err)
	}
	if addr != fk.deriveAddr {
		t.Fatalf("unexpected addr: %v", addr)
	}
	if len(fk.deriveCalls) != 1 {
		t.Fatalf("expected 1 derive call, got %d", len(fk.deriveCalls))
	}
	if !bytes.Equal(u32sToBytes(fk.deriveCalls[0]), u32sToBytes(path)) {
		t.Fatalf("path not forwarded correctly")
	}
}

func TestSignTx_UsesKeystoreSignAndBuildsSignedTx(t *testing.T) {
	fk := &ksStub{signSig: []byte{0xDE, 0xAD, 0xBE, 0xEF}}
	w := New(fk)

	dst := mustAddress(0x11)
	src := mustAddress(0x22)
	txObj := &tx.Tx{
		To:    dst,
		From:  src,
		Value: 123,
	}

	path := keystore.DerivationPath{
		keystore.Hardened(44),
		keystore.Hardened(0),
		keystore.Hardened(0),
		0,
		0,
	}

	stx, err := w.SignTx(txObj, path)
	if err != nil {
		t.Fatalf("SignTx error: %v", err)
	}
	// Signed object checks
	if !bytes.Equal(stx.Sig, fk.signSig) {
		t.Fatalf("signature mismatch")
	}
	// Ensure keystore received the same path and a non-empty hash
	if len(fk.signCalls) != 1 {
		t.Fatalf("expected 1 sign call, got %d", len(fk.signCalls))
	}
	if !bytes.Equal(u32sToBytes(fk.signCalls[0]), u32sToBytes(path)) {
		t.Fatalf("sign path mismatch")
	}
	if len(fk.signMsgs) != 1 || len(fk.signMsgs[0]) == 0 {
		t.Fatalf("expected non-empty tx hash passed to keystore")
	}
	if reflect.DeepEqual(stx, signedtx.SignedTx{}) {
		t.Fatalf("empty SignedTx")
	}
}

func TestNextPeekSetIndex(t *testing.T) {
	w := New(&ksStub{})

	// defaults start at 0
	if got := w.PeekIndex(0, 0); got != 0 {
		t.Fatalf("PeekIndex initial = %d, want 0", got)
	}

	// Next increments
	if got := w.NextIndex(0, 0); got != 0 {
		t.Fatalf("NextIndex first = %d, want 0", got)
	}
	if got := w.PeekIndex(0, 0); got != 1 {
		t.Fatalf("PeekIndex after Next = %d, want 1", got)
	}

	// SetIndex sets next
	w.SetIndex(1, 0, 10)
	if got := w.PeekIndex(1, 0); got != 10 {
		t.Fatalf("PeekIndex account1 = %d, want 10", got)
	}
	if got := w.NextIndex(1, 0); got != 10 {
		t.Fatalf("NextIndex account1 first = %d, want 10", got)
	}
	if got := w.PeekIndex(1, 0); got != 11 {
		t.Fatalf("PeekIndex account1 after Next = %d, want 11", got)
	}
}

func TestNewReceivingAddress_UsesAndIncrementsIndex(t *testing.T) {
	w := New(&ksStub{})

	// First address should use index 0 (stub encodes index in last byte)
	addr1, err := w.NewReceivingAddress()
	if err != nil {
		t.Fatalf("NewReceivingAddress error: %v", err)
	}
	if addr1[19] != 0x00 {
		t.Fatalf("addr1 last byte = %x, want 00", addr1[19])
	}

	// Second address should use index 1
	addr2, err := w.NewReceivingAddress()
	if err != nil {
		t.Fatalf("NewReceivingAddress error: %v", err)
	}
	if addr2[19] != 0x01 {
		t.Fatalf("addr2 last byte = %x, want 01", addr2[19])
	}
}

func TestGetDerivationPath_BIP44Like(t *testing.T) {
	w := New(&ksStub{})
	h := func(x uint32) uint32 { return x | 0x80000000 }

	path := w.GetDerivationPath(0, 2, 1, 7)
	if len(path) != 5 {
		t.Fatalf("path len=%d, want 5", len(path))
	}
	if path[0] != h(44) || path[1] != h(0) || path[2] != h(2) || path[3] != 1 || path[4] != 7 {
		t.Fatalf("unexpected path: %#v", path)
	}
}

func TestNewAddressFor_ReturnsAddrAndIndex(t *testing.T) {
	w := New(&ksStub{})

	addr, idx, err := w.NewAddressFor(0, 0)
	if err != nil {
		t.Fatalf("NewAddressFor error: %v", err)
	}
	if idx != 0 {
		t.Fatalf("idx=%d, want 0", idx)
	}
	if addr[19] != 0x00 {
		t.Fatalf("addr last byte = %x, want 00", addr[19])
	}

	addr2, idx2, err := w.NewAddressFor(0, 0)
	if err != nil {
		t.Fatalf("NewAddressFor(2) error: %v", err)
	}
	if idx2 != 1 {
		t.Fatalf("idx2=%d, want 1", idx2)
	}
	if addr2[19] != 0x01 {
		t.Fatalf("addr2 last byte = %x, want 01", addr2[19])
	}
}
