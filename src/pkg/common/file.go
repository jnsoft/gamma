package common

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetFullPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("cannot get absolute path: %w", err)
	}
	return absPath, nil
}

func IsWritableDir(path string) error {
	if !DirExists(path) {
		return fmt.Errorf("data directory does not exist: %s", path)
	}

	// Try to create and remove a temp file to verify write permissions.
	test := filepath.Join(path, ".writable_check.tmp")
	f, err := os.OpenFile(test, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("data directory not writable: %s", path)
	}
	_ = f.Close()
	_ = os.Remove(test)
	return nil
}

func DirExists(path string) bool {
	// Normalize and resolve symlinks
	clean := filepath.Clean(path)
	resolved, err := filepath.EvalSymlinks(clean)
	if err != nil {
		return false
	}

	// Lstat to inspect the final node without following further links
	fi, err := os.Lstat(resolved)
	if err != nil {
		return false
	}

	// Reject symlinks for directory checks
	if fi.Mode()&os.ModeSymlink != 0 {
		return false
	}
	// Must be a directory
	return fi.IsDir()
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
