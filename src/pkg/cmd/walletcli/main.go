package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"github.com/jnsoft/gamma/src/pkg/domain/address"
	"github.com/jnsoft/gamma/src/pkg/domain/tx"
	"github.com/jnsoft/gamma/src/pkg/keystore"
	"github.com/jnsoft/gamma/src/pkg/wallet"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	switch os.Args[1] {
	case "create":
		cmdCreate()
	case "address":
		cmdAddress()
	case "sign":
		cmdSign()
	default:
		usage()
	}
}

func usage() {
	fmt.Println("walletcli <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create   - create a new wallet")
	fmt.Println("  address  - generate a new receiving address")
	fmt.Println("  sign     - sign a transaction")
}

func cmdCreate() {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	file := fs.String("file", "wallet.json", "keystore file")
	password := fs.String("password", "", "wallet password")
	fs.Parse(os.Args[2:])

	if *password == "" {
		fmt.Println("password required")
		return
	}

	ks, err := keystore.NewFileKeystore([]byte(*password))
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	if err := ks.Save(*file); err != nil {
		fmt.Println("error:", err)
		return
	}

	stateFile := statePathFor(*file)
	w := wallet.New(ks)        // ks is not used here yet, but we initialize state
	_ = w.SaveState(stateFile) // ignore error for first run
	fmt.Println("wallet created:", *file)
}

func cmdAddress() {
	fs := flag.NewFlagSet("address", flag.ExitOnError)
	file := fs.String("file", "wallet.json", "keystore file")
	password := fs.String("password", "", "wallet password")
	account := fs.Uint("account", 0, "account number (hardened)")
	change := fs.Uint("change", 0, "0=external, 1=internal")
	fs.Parse(os.Args[2:])

	ks, err := keystore.LoadFileKeystore(*file)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	if err := ks.Unlock([]byte(*password)); err != nil {
		fmt.Println("invalid password")
		return
	}

	w := wallet.New(ks)
	stateFile := statePathFor(*file)
	if err := w.LoadState(stateFile); err != nil {
		fmt.Println("error:", err)
		return
	}

	addr, idx, err := w.NewAddressFor(uint32(*account), uint32(*change))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	if err := w.SaveState(stateFile); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("new address (account %d, change %d, index %d): %s\n", *account, *change, idx, addr.String())
}

func cmdSign() {
	fs := flag.NewFlagSet("sign", flag.ExitOnError)
	file := fs.String("file", "wallet.json", "keystore file")
	password := fs.String("password", "", "wallet password")
	to := fs.String("to", "", "destination address (hex)")
	amount := fs.Uint("amount", 0, "amount")
	account := fs.Uint("account", 0, "account number (hardened)")
	change := fs.Uint("change", 0, "0=external, 1=internal")
	fs.Parse(os.Args[2:])

	if *to == "" {
		fmt.Println("destination required")
		return
	}

	toAddr, err := hex.DecodeString(*to)
	if err != nil {
		fmt.Println("invalid address")
		return
	}

	validatedToAddress, err := address.ToAddress(toAddr)
	if err != nil {
		fmt.Println("invalid address")
		return
	}

	ks, err := keystore.LoadFileKeystore(*file)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	if err = ks.Unlock([]byte(*password)); err != nil {
		fmt.Println("invalid password")
		return
	}

	w := wallet.New(ks)
	stateFile := statePathFor(*file)
	_ = w.LoadState(stateFile) // optional for index-aware flows

	// Use current index without incrementing for signing
	idx := w.PeekIndex(uint32(*account), uint32(*change))
	path := w.GetDerivationPath(0, uint32(*account), uint32(*change), idx)

	// Derive the From address from the same path
	fromAddr, err := w.NewReceivingAddressWithPath(path)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	tx := &tx.Tx{
		To:    validatedToAddress,
		From:  fromAddr,
		Value: *amount,
	}

	stx, err := w.SignTx(tx, path)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println("transaction signed")
	fmt.Println("signature:", hex.EncodeToString(stx.Sig))
}

func statePathFor(keystorePath string) string {
	return keystorePath + ".state.json"
}
