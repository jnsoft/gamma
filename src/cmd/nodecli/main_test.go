package main

import (
	"os/exec"
	"path/filepath"
	"testing"
)

var cliDir = filepath.Join("/workspace", "src", "cmd", "nodecli")

func TestInit(t *testing.T) {
	tmp := t.TempDir()
	dataDir := filepath.Join(tmp, "data")

	dataDir = ""

	// init
	{
		cmd := exec.Command("go", "run", ".", "init", "-d", dataDir)
		cmd.Dir = cliDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("init failed: %v", err)
		}
	}
}
