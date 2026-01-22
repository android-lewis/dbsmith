package db

import (
	"testing"

	"github.com/android-lewis/dbsmith/internal/models"
)

func TestNewDriverFactory(t *testing.T) {
	f := NewDriverFactory()

	if f == nil {
		t.Fatal("Expected factory instance")
	}

	// Verify default drivers are registered
	if !f.IsSupported("postgres") {
		t.Error("Expected postgres to be supported")
	}

	if !f.IsSupported("mysql") {
		t.Error("Expected mysql to be supported")
	}

	if !f.IsSupported("sqlite") {
		t.Error("Expected sqlite to be supported")
	}
}

func TestFactoryCreate(t *testing.T) {
	f := NewDriverFactory()

	tests := []struct {
		name    string
		conn    *models.Connection
		wantErr bool
		wantNil bool
	}{
		{
			name:    "postgres connection",
			conn:    &models.Connection{Type: models.PostgresType, Host: "localhost"},
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "postgresql alias",
			conn:    &models.Connection{Type: "postgresql", Host: "localhost"},
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "mysql connection",
			conn:    &models.Connection{Type: models.MySQLType, Host: "localhost"},
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "sqlite connection",
			conn:    &models.Connection{Type: models.SQLiteType, Database: "test.db"},
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "nil connection",
			conn:    nil,
			wantErr: true,
			wantNil: true,
		},
		{
			name:    "unsupported type",
			conn:    &models.Connection{Type: "unsupported"},
			wantErr: true,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver, err := f.Create(tt.conn)

			if tt.wantErr && err == nil {
				t.Error("Expected error")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.wantNil && driver != nil {
				t.Error("Expected nil driver")
			}

			if !tt.wantNil && driver == nil {
				t.Error("Expected non-nil driver")
			}
		})
	}
}

func TestFactoryRegister(t *testing.T) {
	f := NewDriverFactory()

	err := f.Register("custom", func() Driver {
		return &MockDriver{}
	})

	if err != nil {
		t.Fatalf("Failed to register custom driver: %v", err)
	}

	if !f.IsSupported("custom") {
		t.Error("Expected custom to be supported")
	}

	driver, err := f.Create(&models.Connection{Type: "custom"})
	if err != nil {
		t.Fatalf("Failed to create custom driver: %v", err)
	}

	if driver == nil {
		t.Error("Expected non-nil driver")
	}

	_, ok := driver.(*MockDriver)
	if !ok {
		t.Error("Expected MockDriver type")
	}
}

func TestFactoryRegisterErrors(t *testing.T) {
	f := NewDriverFactory()

	err := f.Register("", func() Driver {
		return nil
	})
	if err == nil {
		t.Error("Expected error for empty type")
	}

	err = f.Register("test", nil)
	if err == nil {
		t.Error("Expected error for nil constructor")
	}
}

func TestFactoryIsSupported(t *testing.T) {
	f := NewDriverFactory()

	tests := []struct {
		dbType      string
		wantSupport bool
	}{
		{"postgres", true},
		{"postgresql", true},
		{"mysql", true},
		{"mariadb", true},
		{"sqlite", true},
		{"unsupported", false},
		{"", false},
	}

	for _, tt := range tests {
		if f.IsSupported(tt.dbType) != tt.wantSupport {
			t.Errorf("IsSupported(%q) = %v, want %v", tt.dbType, f.IsSupported(tt.dbType), tt.wantSupport)
		}
	}
}

func TestFactoryGetSupportedTypes(t *testing.T) {
	f := NewDriverFactory()

	types := f.GetSupportedTypes()

	if len(types) == 0 {
		t.Error("Expected at least one supported type")
	}

	supportedTypes := make(map[string]bool)
	for _, t := range types {
		supportedTypes[t] = true
	}

	expectedTypes := []string{"postgres", "mysql", "sqlite"}
	for _, expectedType := range expectedTypes {
		if !supportedTypes[expectedType] {
			t.Errorf("Expected %s to be in supported types", expectedType)
		}
	}
}

func TestFactoryCreateTypeErrors(t *testing.T) {
	f := NewDriverFactory()

	// Create with connection that has no type
	conn := &models.Connection{Host: "localhost"}
	_, err := f.Create(conn)
	if err != ErrInvalidConnection {
		t.Errorf("Expected ErrInvalidConnection, got %v", err)
	}
}
