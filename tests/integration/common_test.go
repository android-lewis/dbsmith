//go:build integration
// +build integration

package integration

import (
	"testing"
)

// =============================================================================
// Secrets Manager Tests
// =============================================================================

func TestSecretsManagerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setup := SetupSQLite(t) // SQLite is always available
	defer setup.Close()

	secretsMgr := setup.SecretsManager()

	testCases := []struct {
		name   string
		keyID  string
		secret string
	}{
		{"simple_password", "test-key-1", "password123"},
		{"complex_credential", "test-key-2", "p@ssw0rd!#$%^&*()"},
		{"long_secret", "test-key-3", "verylongpasswordwithrandomcharacters1234567890!@#$%^&*()"},
		{"empty_string", "test-key-4", ""},
		{"unicode_password", "test-key-5", "密码パスワード"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := secretsMgr.StoreSecret(tc.keyID, tc.secret); err != nil {
				t.Fatalf("Failed to store secret: %v", err)
			}

			retrieved, err := secretsMgr.RetrieveSecret(tc.keyID)
			if err != nil {
				t.Fatalf("Failed to retrieve secret: %v", err)
			}

			if retrieved != tc.secret {
				t.Errorf("Secret mismatch: expected %s, got %s", tc.secret, retrieved)
			}
		})
	}
}

// =============================================================================
// CI Detection Tests
// =============================================================================

func TestIsCIDetection(t *testing.T) {
	// This test just verifies the function doesn't panic
	// The actual value depends on the environment
	isCI := IsCI()
	t.Logf("Running in CI: %v", isCI)
}
