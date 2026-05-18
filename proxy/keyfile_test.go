package proxy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadKeyFile(t *testing.T) {
	dir := t.TempDir()
	original := keyFilePath
	keyFilePath = func() string { return filepath.Join(dir, "key.json") }
	defer func() { keyFilePath = original }()

	err := SaveKeyFile("sk-test-key-123")
	if err != nil {
		t.Fatalf("SaveKeyFile failed: %v", err)
	}

	key, err := LoadKeyFile()
	if err != nil {
		t.Fatalf("LoadKeyFile failed: %v", err)
	}
	if key != "sk-test-key-123" {
		t.Errorf("expected sk-test-key-123, got %s", key)
	}
}

func TestLoadKeyFile_NotFound(t *testing.T) {
	dir := t.TempDir()
	original := keyFilePath
	keyFilePath = func() string { return filepath.Join(dir, "nonexistent.json") }
	defer func() { keyFilePath = original }()

	_, err := LoadKeyFile()
	if err == nil {
		t.Error("expected error for missing key file")
	}
}

func TestSaveKeyFile_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "subdir", "nested")
	original := keyFilePath
	keyFilePath = func() string { return filepath.Join(nested, "key.json") }
	defer func() { keyFilePath = original }()

	err := SaveKeyFile("sk-test")
	if err != nil {
		t.Fatalf("SaveKeyFile failed: %v", err)
	}

	if _, err := os.Stat(nested); os.IsNotExist(err) {
		t.Error("directory should have been created")
	}
}
