package secrets

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/android-lewis/dbsmith/internal/security"
)

type EncryptedFileManager struct {
	configDir string
}

type secretsFile struct {
	Secrets map[string]string `json:"secrets"`
}

func (efm *EncryptedFileManager) StoreSecret(keyID, secret string) error {
	if err := os.MkdirAll(efm.configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	keyFile := filepath.Join(efm.configDir, ".secrets.key")
	key, err := efm.getOrCreateEncryptionKey(keyFile)
	if err != nil {
		return err
	}

	enc, err := security.NewEncryptor(key)
	if err != nil {
		return err
	}

	encrypted, err := enc.Encrypt(secret)
	if err != nil {
		return err
	}

	secretFile := filepath.Join(efm.configDir, ".secrets")
	sf, err := efm.loadSecretsFile(secretFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if sf == nil {
		sf = &secretsFile{Secrets: make(map[string]string)}
	}

	sf.Secrets[keyID] = encrypted

	return efm.saveSecretsFile(secretFile, sf)
}

func (efm *EncryptedFileManager) RetrieveSecret(keyID string) (string, error) {
	keyFile := filepath.Join(efm.configDir, ".secrets.key")
	key, err := efm.getOrCreateEncryptionKey(keyFile)
	if err != nil {
		return "", err
	}

	secretFile := filepath.Join(efm.configDir, ".secrets")
	sf, err := efm.loadSecretsFile(secretFile)
	if err != nil {
		return "", fmt.Errorf("failed to load secrets: %w", err)
	}

	if sf == nil {
		return "", fmt.Errorf("secret not found: %s", keyID)
	}

	encrypted, ok := sf.Secrets[keyID]
	if !ok {
		return "", fmt.Errorf("secret not found: %s", keyID)
	}

	enc, err := security.NewEncryptor(key)
	if err != nil {
		return "", err
	}

	return enc.Decrypt(encrypted)
}

func (efm *EncryptedFileManager) DeleteSecret(keyID string) error {
	secretFile := filepath.Join(efm.configDir, ".secrets")
	sf, err := efm.loadSecretsFile(secretFile)
	if err != nil {
		return err
	}

	if sf == nil {
		return fmt.Errorf("secret not found: %s", keyID)
	}

	delete(sf.Secrets, keyID)
	return efm.saveSecretsFile(secretFile, sf)
}

func (efm *EncryptedFileManager) getOrCreateEncryptionKey(keyFile string) ([]byte, error) {
	if data, err := os.ReadFile(keyFile); err == nil && len(data) == 32 {
		return data, nil
	}

	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	if err := os.WriteFile(keyFile, key, 0600); err != nil {
		return nil, fmt.Errorf("failed to save encryption key: %w", err)
	}

	return key, nil
}

func (efm *EncryptedFileManager) loadSecretsFile(secretFile string) (*secretsFile, error) {
	data, err := os.ReadFile(secretFile)
	if err != nil {
		return nil, err
	}

	sf := &secretsFile{}
	if err := json.Unmarshal(data, sf); err != nil {
		return nil, err
	}

	return sf, nil
}

func (efm *EncryptedFileManager) saveSecretsFile(secretFile string, sf *secretsFile) error {
	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(secretFile, data, 0600)
}
