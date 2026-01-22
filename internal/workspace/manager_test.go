package workspace

import (
	"os"
	"testing"

	"github.com/android-lewis/dbsmith/internal/models"
)

func TestManagerNew(t *testing.T) {
	m := New()

	if m.workspace == nil {
		t.Fatal("Workspace should not be nil")
	}

	if m.workspace.Name != "" {
		t.Errorf("Expected empty name, got '%s'", m.workspace.Name)
	}

	if len(m.workspace.Connections) != 0 {
		t.Errorf("Expected 0 connections, got %d", len(m.workspace.Connections))
	}
}

func TestAddConnection(t *testing.T) {
	m := New()

	conn := models.Connection{
		Name: "test-pg",
		Type: models.PostgresType,
		Host: "localhost",
	}

	err := m.AddConnection(conn)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	if len(m.ListConnections()) != 1 {
		t.Error("Connection was not added")
	}
}

func TestAddConnectionDuplicate(t *testing.T) {
	m := New()

	conn := models.Connection{
		Name: "test-pg",
		Type: models.PostgresType,
	}

	_ = m.AddConnection(conn)

	err := m.AddConnection(conn)
	if err == nil {
		t.Error("Expected error when adding duplicate connection")
	}
}

func TestDeleteConnection(t *testing.T) {
	m := New()

	_ = m.AddConnection(models.Connection{Name: "test-pg", Type: models.PostgresType})

	err := m.DeleteConnection("test-pg")
	if err != nil {
		t.Fatalf("Failed to delete connection: %v", err)
	}

	if len(m.ListConnections()) != 0 {
		t.Error("Connection was not deleted")
	}
}

func TestGetConnection(t *testing.T) {
	m := New()

	original := models.Connection{
		Name: "test-pg",
		Type: models.PostgresType,
		Host: "localhost",
		Port: 5432,
	}

	_ = m.AddConnection(original)

	conn, err := m.GetConnection("test-pg")
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
	}

	if conn.Name != original.Name {
		t.Errorf("Expected name '%s', got '%s'", original.Name, conn.Name)
	}

	if conn.Port != original.Port {
		t.Errorf("Expected port %d, got %d", original.Port, conn.Port)
	}
}

func TestAddSavedQuery(t *testing.T) {
	m := New()

	query := models.SavedQuery{
		ID:   "query-1",
		Name: "List Users",
		SQL:  "SELECT * FROM users",
	}

	err := m.AddSavedQuery(query)
	if err != nil {
		t.Fatalf("Failed to add query: %v", err)
	}

	if len(m.ListSavedQueries()) != 1 {
		t.Error("Query was not added")
	}
}

func TestSearchSavedQueries(t *testing.T) {
	m := New()

	_ = m.AddSavedQuery(models.SavedQuery{
		ID:   "q1",
		Name: "List Users",
	})

	_ = m.AddSavedQuery(models.SavedQuery{
		ID:   "q2",
		Name: "List Posts",
	})

	// Search by name
	results := m.SearchSavedQueries("Users")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'Users', got %d", len(results))
	}

	// Search by name (case insensitive)
	results = m.SearchSavedQueries("posts")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'posts', got %d", len(results))
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "workspace-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()

	// Create and save workspace
	m1 := New()
	m1.SetName("test-workspace")
	_ = m1.AddConnection(models.Connection{
		Name: "test-pg",
		Type: models.PostgresType,
		Host: "localhost",
	})

	err = m1.Save(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to save workspace: %v", err)
	}

	// Load workspace
	m2, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load workspace: %v", err)
	}

	if m2.GetName() != "test-workspace" {
		t.Errorf("Expected name 'test-workspace', got '%s'", m2.GetName())
	}

	conns := m2.ListConnections()
	if len(conns) != 1 {
		t.Errorf("Expected 1 connection, got %d", len(conns))
	}

	if conns[0].Name != "test-pg" {
		t.Errorf("Expected connection 'test-pg', got '%s'", conns[0].Name)
	}
}

func TestUpdateConnection(t *testing.T) {
	m := New()

	original := models.Connection{
		Name: "test-pg",
		Type: models.PostgresType,
		Host: "localhost",
	}

	_ = m.AddConnection(original)

	updated := original
	updated.Host = "remote.example.com"

	err := m.UpdateConnection(updated)
	if err != nil {
		t.Fatalf("Failed to update connection: %v", err)
	}

	conn, _ := m.GetConnection("test-pg")
	if conn.Host != "remote.example.com" {
		t.Errorf("Expected host 'remote.example.com', got '%s'", conn.Host)
	}
}

func TestAutocompletePreferences(t *testing.T) {
	t.Run("default enabled on new workspace", func(t *testing.T) {
		m := New()
		if !m.GetAutocompleteEnabled() {
			t.Error("Expected autocomplete to be enabled by default")
		}
	})

	t.Run("set autocomplete disabled", func(t *testing.T) {
		m := New()
		err := m.SetAutocompleteEnabled(false)
		if err != nil {
			t.Fatalf("Failed to set autocomplete: %v", err)
		}
		if m.GetAutocompleteEnabled() {
			t.Error("Expected autocomplete to be disabled")
		}
	})

	t.Run("set autocomplete enabled", func(t *testing.T) {
		m := New()
		_ = m.SetAutocompleteEnabled(false)
		err := m.SetAutocompleteEnabled(true)
		if err != nil {
			t.Fatalf("Failed to set autocomplete: %v", err)
		}
		if !m.GetAutocompleteEnabled() {
			t.Error("Expected autocomplete to be enabled")
		}
	})

	t.Run("persistence across save/load", func(t *testing.T) {
		tmpfile, err := os.CreateTemp("", "workspace-prefs-*.yaml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer func() {
			_ = os.Remove(tmpfile.Name())
		}()

		// Create workspace with autocomplete disabled
		m1 := New()
		m1.SetName("pref-test")
		_ = m1.SetAutocompleteEnabled(false)
		err = m1.Save(tmpfile.Name())
		if err != nil {
			t.Fatalf("Failed to save workspace: %v", err)
		}

		// Load and verify
		m2, err := Load(tmpfile.Name())
		if err != nil {
			t.Fatalf("Failed to load workspace: %v", err)
		}

		if m2.GetAutocompleteEnabled() {
			t.Error("Expected autocomplete to be disabled after load")
		}
	})

	t.Run("backwards compatibility - load workspace without preferences", func(t *testing.T) {
		tmpfile, err := os.CreateTemp("", "workspace-legacy-*.yaml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer func() {
			_ = os.Remove(tmpfile.Name())
		}()

		// Create a workspace YAML without preferences field (simulating old format)
		legacyYAML := `name: legacy-workspace
connections: []
saved_queries: []
version: 1
`
		err = os.WriteFile(tmpfile.Name(), []byte(legacyYAML), 0600)
		if err != nil {
			t.Fatalf("Failed to write legacy YAML: %v", err)
		}

		// Load and verify defaults are applied
		m, err := Load(tmpfile.Name())
		if err != nil {
			t.Fatalf("Failed to load legacy workspace: %v", err)
		}

		if !m.GetAutocompleteEnabled() {
			t.Error("Expected autocomplete to be enabled by default for legacy workspace")
		}
	})
}
