package secrets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptedFileManager(t *testing.T) {
	tmpDir := t.TempDir()

	manager := &EncryptedFileManager{configDir: tmpDir}

	err := manager.StoreSecret("test-key", "test-secret")
	if err != nil {
		t.Fatalf("Failed to store secret: %v", err)
	}

	secret, err := manager.RetrieveSecret("test-key")
	if err != nil {
		t.Fatalf("Failed to retrieve secret: %v", err)
	}

	if secret != "test-secret" {
		t.Errorf("Expected 'test-secret', got '%s'", secret)
	}

	_, err = manager.RetrieveSecret("non-existent")
	if err == nil {
		t.Error("Expected error when retrieving non-existent secret")
	}

	err = manager.DeleteSecret("test-key")
	if err != nil {
		t.Fatalf("Failed to delete secret: %v", err)
	}

	_, err = manager.RetrieveSecret("test-key")
	if err == nil {
		t.Error("Expected error when retrieving deleted secret")
	}
}

func TestEncryptionKeyGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	manager := &EncryptedFileManager{configDir: tmpDir}

	keyFile := filepath.Join(tmpDir, ".secrets.key")

	key1, err := manager.getOrCreateEncryptionKey(keyFile)
	if err != nil {
		t.Fatalf("Failed to create key: %v", err)
	}

	if len(key1) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key1))
	}

	info, err := os.Stat(keyFile)
	if err != nil {
		t.Fatalf("Key file not created: %v", err)
	}

	if !info.Mode().IsRegular() {
		t.Error("Key file is not a regular file")
	}

	key2, err := manager.getOrCreateEncryptionKey(keyFile)
	if err != nil {
		t.Fatalf("Failed to load key: %v", err)
	}

	for i := range key1 {
		if key1[i] != key2[i] {
			t.Error("Loaded key doesn't match original")
			break
		}
	}
}
