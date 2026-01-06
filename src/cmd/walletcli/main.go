package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jnsoft/gamma/src/pkg/crypto"
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
	case "passwd":
		cmdPasswd()
	case "export":
		cmdExportSeed()
	case "import":
		cmdImportSeed()
	case "mnemonic-export":
		cmdMnemonicExport()
	case "mnemonic-import":
		cmdMnemonicImport()
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
	fmt.Println("  passwd   - change wallet password")
	fmt.Println("  export   - export wallet seed")
	fmt.Println("  import   - import wallet seed")
	fmt.Println("  mnemonic-export  - export 24-word mnemonic")
	fmt.Println("  mnemonic-import  - import 24-word mnemonic")
	fmt.Println("  init     - initialize data directory")
}

func cmdCreate() {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	file := fs.String("file", "wallet.json", "keystore file")
	password := fs.String("password", "", "wallet password")
	fs.Parse(os.Args[2:])

	if *password == "" {
		fmt.Println("password required")
		os.Exit(1)
	}

	if _, err := os.Stat(*file); err == nil {
		fmt.Println("error: file already exists:", *file)
		os.Exit(1)
	}

	ks, err := keystore.NewFileKeystore([]byte(*password))
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if err := ks.Save(*file); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
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
		os.Exit(1)
	}
	if err := ks.Unlock([]byte(*password)); err != nil {
		fmt.Println("invalid password")
		os.Exit(1)
	}

	w := wallet.New(ks)
	stateFile := statePathFor(*file)
	if err := w.LoadState(stateFile); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	addr, idx, err := w.NewAddressFor(uint32(*account), uint32(*change))
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	if err := w.SaveState(stateFile); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
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
		os.Exit(1)
	}

	toAddr, err := hex.DecodeString(*to)
	if err != nil {
		fmt.Println("invalid address")
		os.Exit(1)
	}

	validatedToAddress, err := address.ToAddress(toAddr)
	if err != nil {
		fmt.Println("invalid address")
		os.Exit(1)
	}

	ks, err := keystore.LoadFileKeystore(*file)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if err = ks.Unlock([]byte(*password)); err != nil {
		fmt.Println("invalid password")
		os.Exit(1)
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
		os.Exit(1)
	}

	tx := &tx.Tx{
		To:    validatedToAddress,
		From:  fromAddr,
		Value: *amount,
	}

	stx, err := w.SignTx(tx, path)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("transaction signed")
	fmt.Println("signature:", hex.EncodeToString(stx.Sig))
}

func cmdPasswd() {
	fs := flag.NewFlagSet("passwd", flag.ExitOnError)
	file := fs.String("file", "wallet.json", "keystore file")
	old := fs.String("old", "", "old password")
	new := fs.String("new", "", "new password")
	fs.Parse(os.Args[2:])
	ks, err := keystore.LoadFileKeystore(*file)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	if err := ks.Unlock([]byte(*old)); err != nil {
		fmt.Println("invalid password")
		os.Exit(1)
	}
	if err := ks.ChangePassword([]byte(*old), []byte(*new)); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	if err := ks.Save(*file); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println("password updated")
}

func cmdExportSeed() {
	fs := flag.NewFlagSet("export-seed", flag.ExitOnError)
	file := fs.String("file", "wallet.json", "keystore file")
	password := fs.String("password", "", "wallet password")
	out := fs.String("out", "wallet.seed", "output file for raw seed (0600)")
	fs.Parse(os.Args[2:])

	ks, err := keystore.LoadFileKeystore(*file)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	if err := ks.Unlock([]byte(*password)); err != nil {
		fmt.Println("invalid password")
		os.Exit(1)
	}
	seed, err := ks.ExportSeed()
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	seedHex := hex.EncodeToString(seed)
	if err := os.WriteFile(*out, []byte(seedHex+"\n"), 0600); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println("seed exported to:", *out)
}

func cmdImportSeed() {
	fs := flag.NewFlagSet("import-seed", flag.ExitOnError)
	file := fs.String("file", "wallet.json", "keystore file")
	password := fs.String("password", "", "wallet password")
	seedHex := fs.String("seed", "", "seed hex")
	fs.Parse(os.Args[2:])
	seed, err := hex.DecodeString(*seedHex)
	if err != nil || len(seed) == 0 {
		fmt.Println("invalid seed")
		os.Exit(1)
	}
	ks, err := keystore.NewFileKeystoreFromSeed(seed, []byte(*password))
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	if err := ks.Save(*file); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	// init state file
	_ = wallet.New(ks).SaveState(statePathFor(*file))
	fmt.Println("seed imported:", *file)
}

func cmdMnemonicExport() {
	fs := flag.NewFlagSet("mnemonic-export", flag.ExitOnError)
	file := fs.String("file", "wallet.json", "keystore file")
	password := fs.String("password", "", "wallet password")
	out := fs.String("out", "wallet.mnemonic", "output file for mnemonic (0600)")
	fs.Parse(os.Args[2:])

	ks, err := keystore.LoadFileKeystore(*file)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	if err := ks.Unlock([]byte(*password)); err != nil {
		fmt.Println("invalid password")
		os.Exit(1)
	}

	seed, err := ks.ExportSeed()
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	mn, err := crypto.GetMnemonic(seed)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	// Write mnemonic to file with restrictive perms
	if err := os.WriteFile(*out, []byte(mn+"\n"), 0600); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println("mnemonic exported to:", *out)
}

func cmdMnemonicImport() {
	fs := flag.NewFlagSet("mnemonic-import", flag.ExitOnError)
	file := fs.String("file", "wallet.json", "keystore file")
	password := fs.String("password", "", "wallet password")
	mn := fs.String("mnemonic", "", "24-word mnemonic")
	fs.Parse(os.Args[2:])

	if *mn == "" {
		fmt.Println("mnemonic required")
		os.Exit(1)
	}

	seed, err := crypto.GetEntropyFromMnemonic(*mn)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	ks, err := keystore.NewFileKeystoreFromSeed(seed, []byte(*password))
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	if err := ks.Save(*file); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	_ = wallet.New(ks).SaveState(statePathFor(*file))
	fmt.Println("mnemonic imported:", *file)
}

func statePathFor(keystorePath string) string {
	dir := filepath.Dir(keystorePath)
	base := filepath.Base(keystorePath)
	ext := filepath.Ext(base) // ".json"
	name := strings.TrimSuffix(base, ext)
	return filepath.Join(dir, name+".state.json")
}
