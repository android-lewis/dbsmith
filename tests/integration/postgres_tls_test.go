//go:build integration
// +build integration

package integration

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

// =============================================================================
// TLS Connection Tests - Positive Cases
// =============================================================================

// TestPostgresTLSConnectionRequire verifies that sslmode=require works.
// This mode encrypts the connection but doesn't verify the server certificate.
func TestPostgresTLSConnectionRequire(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setup := SetupPostgresTLSContainer(t)
	defer setup.Close()

	// sslmode=require: encrypted connection, no certificate verification
	driver := setup.ConnectWithSSLMode(t, "require", "")
	defer driver.Disconnect(context.Background())

	// Verify connection works
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := driver.Ping(ctx); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	// Verify we're actually using SSL by querying pg_stat_ssl
	result, err := driver.ExecuteQuery(ctx, "SELECT ssl FROM pg_stat_ssl WHERE pid = pg_backend_pid()")
	if err != nil {
		t.Fatalf("Failed to check SSL status: %v", err)
	}

	if len(result.Rows) == 0 {
		t.Fatal("Expected SSL status row, got none")
	}

	sslEnabled, ok := result.Rows[0][0].(bool)
	if !ok || !sslEnabled {
		t.Errorf("Expected SSL connection to be active, got: %v", result.Rows[0][0])
	}

	t.Log("Successfully connected with sslmode=require, SSL is active")
}

// TestPostgresTLSConnectionVerifyCA verifies that sslmode=verify-ca works.
// This mode encrypts the connection and verifies the server certificate
// against the provided CA certificate.
func TestPostgresTLSConnectionVerifyCA(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setup := SetupPostgresTLSContainer(t)
	defer setup.Close()

	// sslmode=verify-ca: encrypted + verify server cert against CA
	driver := setup.ConnectWithSSLMode(t, "verify-ca", setup.CACertPath)
	defer driver.Disconnect(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := driver.Ping(ctx); err != nil {
		t.Fatalf("Ping failed with verify-ca: %v", err)
	}

	// Verify SSL is active
	result, err := driver.ExecuteQuery(ctx, "SELECT ssl FROM pg_stat_ssl WHERE pid = pg_backend_pid()")
	if err != nil {
		t.Fatalf("Failed to check SSL status: %v", err)
	}

	if len(result.Rows) == 0 || result.Rows[0][0] != true {
		t.Error("Expected SSL connection to be active")
	}

	t.Log("Successfully connected with sslmode=verify-ca, certificate verified")
}

// TestPostgresTLSConnectionVerifyFull verifies that sslmode=verify-full works.
// This mode encrypts the connection, verifies the server certificate against
// the CA, and verifies the server hostname matches the certificate.
func TestPostgresTLSConnectionVerifyFull(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setup := SetupPostgresTLSContainer(t)
	defer setup.Close()

	// sslmode=verify-full: encrypted + verify cert + verify hostname
	// The certificate was generated with "localhost" and "127.0.0.1" as SANs
	driver := setup.ConnectWithSSLMode(t, "verify-full", setup.CACertPath)
	defer driver.Disconnect(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := driver.Ping(ctx); err != nil {
		t.Fatalf("Ping failed with verify-full: %v", err)
	}

	t.Log("Successfully connected with sslmode=verify-full, hostname verified")
}

// =============================================================================
// TLS Connection Tests - Negative Cases
// =============================================================================

// TestPostgresTLSConnectionDisableFails verifies that sslmode=disable is rejected.
// The TLS container is configured to reject non-SSL connections.
func TestPostgresTLSConnectionDisableFails(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setup := SetupPostgresTLSContainer(t)
	defer setup.Close()

	// sslmode=disable should fail because server requires SSL (pg_hba.conf)
	err := setup.TryConnectWithSSLMode("disable", "")
	if err == nil {
		t.Fatal("Expected connection to fail with sslmode=disable, but it succeeded")
	}

	t.Logf("Connection correctly rejected with sslmode=disable: %v", err)
}

// TestPostgresTLSConnectionWrongCACertFails verifies that the wrong CA cert is rejected.
// When using verify-ca or verify-full, the server certificate must be signed by
// the provided CA certificate.
func TestPostgresTLSConnectionWrongCACertFails(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setup := SetupPostgresTLSContainer(t)
	defer setup.Close()

	// Generate a different CA cert (won't match server cert)
	wrongCerts := GenerateTLSCerts(t, "wrong-host")

	// sslmode=verify-ca with wrong CA should fail
	err := setup.TryConnectWithSSLMode("verify-ca", wrongCerts.CACertPath)
	if err == nil {
		t.Fatal("Expected connection to fail with wrong CA cert, but it succeeded")
	}

	t.Logf("Connection correctly rejected with wrong CA: %v", err)
}

// TestPostgresTLSConnectionMissingCACertFails verifies that a missing CA cert fails.
// When using verify-ca or verify-full without providing a CA cert path, the
// connection should fail.
func TestPostgresTLSConnectionMissingCACertFails(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	setup := SetupPostgresTLSContainer(t)
	defer setup.Close()

	// sslmode=verify-ca with non-existent CA cert should fail
	nonExistentPath := filepath.Join(t.TempDir(), "nonexistent.crt")
	err := setup.TryConnectWithSSLMode("verify-ca", nonExistentPath)
	if err == nil {
		t.Fatal("Expected connection to fail with missing CA cert, but it succeeded")
	}

	t.Logf("Connection correctly rejected with missing CA: %v", err)
}
