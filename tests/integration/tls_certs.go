package integration

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type TLSCertBundle struct {
	CACertPath     string
	ServerCertPath string
	ServerKeyPath  string
	TempDir        string
}

func GenerateTLSCerts(t *testing.T, serverHostnames ...string) *TLSCertBundle {
	t.Helper()

	tempDir := t.TempDir()

	if len(serverHostnames) == 0 {
		serverHostnames = []string{"localhost", "127.0.0.1"}
	}

	caKey, caCert := generateCA(t)
	caCertPath := filepath.Join(tempDir, "ca.crt")
	writeCertPEM(t, caCertPath, caCert.Raw)

	serverKey, serverCertDER := generateServerCert(t, caKey, caCert, serverHostnames)
	serverCertPath := filepath.Join(tempDir, "server.crt")
	serverKeyPath := filepath.Join(tempDir, "server.key")
	writeCertPEM(t, serverCertPath, serverCertDER)
	writeKeyPEM(t, serverKeyPath, serverKey)

	if err := os.Chmod(serverKeyPath, 0600); err != nil {
		t.Fatalf("Failed to set permissions on server key: %v", err)
	}

	return &TLSCertBundle{
		CACertPath:     caCertPath,
		ServerCertPath: serverCertPath,
		ServerKeyPath:  serverKeyPath,
		TempDir:        tempDir,
	}
}

func generateCA(t *testing.T) (*ecdsa.PrivateKey, *x509.Certificate) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate CA key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"DBSmith Test CA"},
			CommonName:   "DBSmith Test Root CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(1 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
		MaxPathLen:            1,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("Failed to create CA certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse CA certificate: %v", err)
	}

	return key, cert
}

func generateServerCert(t *testing.T, caKey *ecdsa.PrivateKey, caCert *x509.Certificate, hostnames []string) (*ecdsa.PrivateKey, []byte) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate server key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"DBSmith Test"},
			CommonName:   hostnames[0],
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(1 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	for _, h := range hostnames {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, caCert, &key.PublicKey, caKey)
	if err != nil {
		t.Fatalf("Failed to create server certificate: %v", err)
	}

	return key, certDER
}

func writeCertPEM(t *testing.T, path string, derBytes []byte) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create cert file %s: %v", path, err)
	}
	defer f.Close()

	if err := pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		t.Fatalf("Failed to write PEM certificate: %v", err)
	}
}

func writeKeyPEM(t *testing.T, path string, key *ecdsa.PrivateKey) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create key file %s: %v", path, err)
	}
	defer f.Close()

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("Failed to marshal private key: %v", err)
	}

	if err := pem.Encode(f, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}); err != nil {
		t.Fatalf("Failed to write PEM key: %v", err)
	}
}
