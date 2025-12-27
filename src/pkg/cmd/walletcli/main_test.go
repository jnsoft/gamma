package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

var cliDir = filepath.Join("/workspace", "src", "pkg", "cmd", "walletcli")
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
