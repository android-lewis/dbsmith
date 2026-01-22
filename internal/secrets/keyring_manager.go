package secrets

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

type KeyringManager struct {
}

func (km *KeyringManager) StoreSecret(keyID, secret string) error {
	err := keyring.Set("dbsmith", keyID, secret)
	if err != nil {
		return fmt.Errorf("failed to store secret: %w", err)
	}

	return nil
}

func (km *KeyringManager) RetrieveSecret(keyID string) (string, error) {
	secret, err := keyring.Get("dbsmith", keyID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret: %w", err)
	}

	return secret, nil
}

func (km *KeyringManager) DeleteSecret(keyID string) error {
	err := keyring.Delete("dbsmith", keyID)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}
