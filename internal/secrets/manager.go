package secrets

import (
	"github.com/zalando/go-keyring"
)

type Manager interface {
	StoreSecret(keyID, secret string) error
	RetrieveSecret(keyID string) (string, error)
	DeleteSecret(keyID string) error
}

func NewManager(configDir string) (Manager, error) {
	testErr := keyring.Set("dbsmith-test", "test-key", "test-value")
	if testErr == nil {
		_ = keyring.Delete("dbsmith-test", "test-key")
		return &KeyringManager{}, nil
	}

	return &EncryptedFileManager{configDir: configDir}, nil
}
