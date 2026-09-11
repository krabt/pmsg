package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !FileExists(file) || !FileExists(dir) {
		t.Fatal("FileExists returned false for an existing path")
	}
	if FileExists(filepath.Join(dir, "missing")) {
		t.Fatal("FileExists returned true for a missing path")
	}
}
