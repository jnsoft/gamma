package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

var cliDir = filepath.Join("/workspace", "src", "cmd", "walletcli")
var addrLine = regexp.MustCompile(`new address \(account (\d+), change (\d+), index (\d+)\): ([0-9a-fA-F]+)`)

func TestCreateAndAddress(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "wallet.json")

	// go run create
	{
		cmd := exec.Command("go", "run", ".", "create",
			"-file", file, "-password", "testpass")
		cmd.Dir = cliDir
		var out, errb bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("create failed: %v, stderr=%s", err, errb.String())
		}
	}

	// first address (expect index 0)
	var out1 bytes.Buffer
	{
		cmd := exec.Command("go", "run", ".", "address",
			"-file", file, "-password", "testpass", "-account", "0", "-change", "0")
		cmd.Dir = cliDir
		var errb bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out1, &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("address 1 failed: %v, stderr=%s", err, errb.String())
		}
		m := addrLine.FindSubmatch(out1.Bytes())
		if m == nil {
			t.Fatalf("unexpected output: %s", out1.String())
		}
		if string(m[1]) != "0" || string(m[2]) != "0" || string(m[3]) != "0" {
			t.Fatalf("first address index not 0: %q", m[3])
		}
	}

	// second address (expect index 1)
	var out2 bytes.Buffer
	{
		cmd := exec.Command("go", "run", ".", "address",
			"-file", file, "-password", "testpass", "-account", "0", "-change", "0")
		cmd.Dir = cliDir
		var errb bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out2, &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("address 2 failed: %v, stderr=%s", err, errb.String())
		}
		m := addrLine.FindSubmatch(out2.Bytes())
		if m == nil {
			t.Fatalf("unexpected output: %s", out2.String())
		}
		if string(m[3]) != "1" {
			t.Fatalf("second address index not 1: %q", m[3])
		}
	}
}

func TestSign(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "wallet.json")

	// create wallet
	{
		cmd := exec.Command("go", "run", ".", "create",
			"-file", file, "-password", "testpass")
		cmd.Dir = cliDir
		var errb bytes.Buffer
		cmd.Stderr = &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("create failed: %v, stderr=%s", err, errb.String())
		}
	}

	// derive an address to use as "to"
	var addrOut bytes.Buffer
	{
		cmd := exec.Command("go", "run", ".", "address",
			"-file", file, "-password", "testpass",
			"-account", "0", "-change", "0")
		cmd.Dir = cliDir
		cmd.Stdout = &addrOut
		var errb bytes.Buffer
		cmd.Stderr = &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("address failed: %v, stderr=%s", err, errb.String())
		}
	}
	// parse last token (hex address)
	fields := bytes.Fields(addrOut.Bytes())
	if len(fields) == 0 {
		t.Fatalf("no address output")
	}
	to := string(fields[len(fields)-1])

	// sign
	{
		cmd := exec.Command("go", "run", ".", "sign",
			"-file", file, "-password", "testpass",
			"-to", to, "-amount", "123", "-account", "0", "-change", "0")
		cmd.Dir = cliDir
		var out, errb bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("sign failed: %v, stderr=%s", err, errb.String())
		}
		if !bytes.Contains(out.Bytes(), []byte("transaction signed")) {
			t.Fatalf("unexpected output: %s", out.String())
		}
	}
}

func TestPasswd_ChangePasswordFlow(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "wallet.json")

	// create with old password
	{
		cmd := exec.Command("go", "run", ".", "create", "-file", file, "-password", "oldpass")
		cmd.Dir = cliDir
		var errb bytes.Buffer
		cmd.Stderr = &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("create failed: %v, stderr=%s", err, errb.String())
		}
	}

	// change password to newpass
	{
		cmd := exec.Command("go", "run", ".", "passwd",
			"-file", file, "-old", "oldpass", "-new", "newpass")
		cmd.Dir = cliDir
		var out, errb bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("passwd failed: %v, stderr=%s", err, errb.String())
		}
		if !bytes.Contains(out.Bytes(), []byte("password updated")) {
			t.Fatalf("unexpected output: %s", out.String())
		}
	}

	// address with old password should fail
	{
		cmd := exec.Command("go", "run", ".", "address",
			"-file", file, "-password", "oldpass", "-account", "0", "-change", "0")
		cmd.Dir = cliDir
		var out, errb bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errb
		// expect non-zero exit
		if err := cmd.Run(); err == nil {
			t.Fatalf("address succeeded with old password unexpectedly: %s", out.String())
		}
	}

	// address with new password should succeed and show index 0
	{
		cmd := exec.Command("go", "run", ".", "address",
			"-file", file, "-password", "newpass", "-account", "0", "-change", "0")
		cmd.Dir = cliDir
		var out, errb bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("address failed with new password: %v, stderr=%s", err, errb.String())
		}
		m := addrLine.FindSubmatch(out.Bytes())
		if m == nil || string(m[3]) != "0" {
			t.Fatalf("expected index 0 after password change, got: %s", out.String())
		}
	}
}

func TestSign_UsesPeekIndex_NotIncrement(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "wallet.json")

	// create
	{
		cmd := exec.Command("go", "run", ".", "create", "-file", file, "-password", "testpass")
		cmd.Dir = cliDir
		var errb bytes.Buffer
		cmd.Stderr = &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("create failed: %v, stderr=%s", err, errb.String())
		}
	}

	// derive two addresses to advance index to 2
	for i := 0; i < 2; i++ {
		cmd := exec.Command("go", "run", ".", "address",
			"-file", file, "-password", "testpass", "-account", "0", "-change", "0")
		cmd.Dir = cliDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("address #%d failed: %v", i+1, err)
		}
	}

	// derive one more address; expect index 2
	var out bytes.Buffer
	{
		cmd := exec.Command("go", "run", ".", "address",
			"-file", file, "-password", "testpass", "-account", "0", "-change", "0")
		cmd.Dir = cliDir
		var errb bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("address 3 failed: %v, stderr=%s", err, errb.String())
		}
		m := addrLine.FindSubmatch(out.Bytes())
		if m == nil || string(m[3]) != "2" {
			t.Fatalf("expected index 2, got output: %s", out.String())
		}
		to := string(m[4])

		// sign should not increment index; it uses PeekIndex
		var signOut, signErr bytes.Buffer
		cmd = exec.Command("go", "run", ".", "sign",
			"-file", file, "-password", "testpass",
			"-to", to, "-amount", "123", "-account", "0", "-change", "0")
		cmd.Dir = cliDir
		cmd.Stdout, cmd.Stderr = &signOut, &signErr
		if err := cmd.Run(); err != nil {
			t.Fatalf("sign failed: %v, stderr=%s", err, signErr.String())
		}
		if !bytes.Contains(signOut.Bytes(), []byte("transaction signed")) {
			t.Fatalf("unexpected sign output: %s", signOut.String())
		}

		// derive again; index should now be 3 (since previous address call incremented from 2)
		var outNext bytes.Buffer
		cmd = exec.Command("go", "run", ".", "address",
			"-file", file, "-password", "testpass", "-account", "0", "-change", "0")
		cmd.Dir = cliDir
		cmd.Stdout = &outNext
		if err := cmd.Run(); err != nil {
			t.Fatalf("address next failed: %v", err)
		}
		m2 := addrLine.FindSubmatch(outNext.Bytes())
		if m2 == nil || string(m2[3]) != "3" {
			t.Fatalf("expected index 3 after sign (no increment), got: %s", outNext.String())
		}
	}
}

func TestExportImportSeed_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "wallet.json")
	seedFile := filepath.Join(tmp, "wallet.seed")

	// create wallet
	{
		cmd := exec.Command("go", "run", ".", "create", "-file", file, "-password", "testpass")
		cmd.Dir = cliDir
		var errb bytes.Buffer
		cmd.Stderr = &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("create failed: %v, stderr=%s", err, errb.String())
		}
	}

	// export seed to file
	{
		cmd := exec.Command("go", "run", ".", "export",
			"-file", file, "-password", "testpass", "-out", seedFile)
		cmd.Dir = cliDir
		var out, errb bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("export failed: %v, stderr=%s", err, errb.String())
		}
		if _, err := os.Stat(seedFile); err != nil {
			t.Fatalf("seed file not written: %v", err)
		}
	}

	// import seed into a new wallet file
	file2 := filepath.Join(tmp, "wallet2.json")
	{
		seedHex, err := os.ReadFile(seedFile)
		if err != nil {
			t.Fatalf("read seed: %v", err)
		}
		cmd := exec.Command("go", "run", ".", "import",
			"-file", file2, "-password", "testpass", "-seed", string(bytes.TrimSpace(seedHex)))
		cmd.Dir = cliDir
		var out, errb bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("import failed: %v, stderr=%s", err, errb.String())
		}
		if _, err := os.Stat(file2); err != nil {
			t.Fatalf("imported wallet file missing: %v", err)
		}
	}

	// derive first address from imported wallet -> index 0
	{
		cmd := exec.Command("go", "run", ".", "address",
			"-file", file2, "-password", "testpass", "-account", "0", "-change", "0")
		cmd.Dir = cliDir
		var out, errb bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("address failed: %v, stderr=%s", err, errb.String())
		}
		m := addrLine.FindSubmatch(out.Bytes())
		if m == nil || string(m[3]) != "0" {
			t.Fatalf("expected index 0 after import, got: %s", out.String())
		}
	}
}

func TestMnemonicExportImport_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "wallet.json")
	mnFile := filepath.Join(tmp, "wallet.mnemonic")

	// create wallet
	{
		cmd := exec.Command("go", "run", ".", "create", "-file", file, "-password", "testpass")
		cmd.Dir = cliDir
		var errb bytes.Buffer
		cmd.Stderr = &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("create failed: %v, stderr=%s", err, errb.String())
		}
	}

	// export mnemonic to file
	{
		cmd := exec.Command("go", "run", ".", "mnemonic-export",
			"-file", file, "-password", "testpass", "-out", mnFile)
		cmd.Dir = cliDir
		var out, errb bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("mnemonic export failed: %v, stderr=%s", err, errb.String())
		}
		if _, err := os.Stat(mnFile); err != nil {
			t.Fatalf("mnemonic file not written: %v", err)
		}
	}

	// import mnemonic into a new wallet file
	file2 := filepath.Join(tmp, "wallet2.json")
	{
		mn, err := os.ReadFile(mnFile)
		if err != nil {
			t.Fatalf("read mnemonic: %v", err)
		}
		cmd := exec.Command("go", "run", ".", "mnemonic-import",
			"-file", file2, "-password", "testpass", "-mnemonic", string(bytes.TrimSpace(mn)))
		cmd.Dir = cliDir
		var out, errb bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("mnemonic import failed: %v, stderr=%s", err, errb.String())
		}
		if _, err := os.Stat(file2); err != nil {
			t.Fatalf("imported wallet file missing: %v", err)
		}
	}

	// address derivation on imported wallet -> index 0
	{
		cmd := exec.Command("go", "run", ".", "address",
			"-file", file2, "-password", "testpass", "-account", "0", "-change", "0")
		cmd.Dir = cliDir
		var out, errb bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("address failed: %v, stderr=%s", err, errb.String())
		}
		m := addrLine.FindSubmatch(out.Bytes())
		if m == nil || string(m[3]) != "0" {
			t.Fatalf("expected index 0 after mnemonic import, got: %s", out.String())
		}
	}
}

func TestMnemonicExport_FilePermissions(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "wallet.json")
	mnFile := filepath.Join(tmp, "wallet.mnemonic")

	// create
	{
		cmd := exec.Command("go", "run", ".", "create", "-file", file, "-password", "testpass")
		cmd.Dir = cliDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("create failed: %v", err)
		}
	}
	// export
	{
		cmd := exec.Command("go", "run", ".", "mnemonic-export",
			"-file", file, "-password", "testpass", "-out", mnFile)
		cmd.Dir = cliDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("mnemonic export failed: %v", err)
		}
	}
	// check mode == 0600 (Alpine Linux)
	info, err := os.Stat(mnFile)
	if err != nil {
		t.Fatalf("stat mnemonic: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("mnemonic file permissions = %o, want 600", info.Mode().Perm())
	}
}
