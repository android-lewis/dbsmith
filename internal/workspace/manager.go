package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/android-lewis/dbsmith/internal/constants"
	"github.com/android-lewis/dbsmith/internal/logging"
	"github.com/android-lewis/dbsmith/internal/models"
	"gopkg.in/yaml.v3"
)

func GetDefaultWorkspacePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	defaultPath := filepath.Join(home, ".config", "dbsmith", "workspace.yaml")
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath
	}
	return ""
}

type Manager struct {
	workspace *models.Workspace
	filePath  string
}

func New() *Manager {
	return &Manager{
		workspace: &models.Workspace{
			Connections:  []models.Connection{},
			SavedQueries: []models.SavedQuery{},
			Preferences:  &models.UserPreferences{AutocompleteEnabled: true},
			CreatedAt:    time.Now(),
			LastModified: time.Now(),
			Version:      constants.WorkspaceVersion,
		},
	}
}

func Load(filePath string) (*Manager, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace file: %w", err)
	}

	ws := &models.Workspace{}
	if err := yaml.Unmarshal(data, ws); err != nil {
		return nil, fmt.Errorf("failed to parse workspace file: %w", err)
	}

	if ws.Preferences == nil {
		ws.Preferences = &models.UserPreferences{AutocompleteEnabled: true}
	}

	return &Manager{
		workspace: ws,
		filePath:  filePath,
	}, nil
}

func (m *Manager) autoSave() error {
	if m.filePath != "" {
		return m.Save(m.filePath)
	}
	return nil
}

func (m *Manager) Save(filePath string) error {
	m.workspace.LastModified = time.Now()
	data, err := yaml.Marshal(m.workspace)
	if err != nil {
		return fmt.Errorf("failed to marshal workspace: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write workspace file: %w", err)
	}

	m.filePath = filePath
	return nil
}

func (m *Manager) SetName(name string) {
	m.workspace.Name = name
	m.workspace.LastModified = time.Now()
}

func (m *Manager) GetName() string {
	return m.workspace.Name
}

func (m *Manager) AddConnection(conn models.Connection) error {
	if conn.Name == "" {
		return errors.New("connection name is required")
	}

	if conn.Type == "" {
		return errors.New("connection type is required")
	}

	for _, c := range m.workspace.Connections {
		if c.Name == conn.Name {
			return fmt.Errorf("connection with name '%s' already exists", conn.Name)
		}
	}

	conn.CreatedAt = time.Now()
	conn.LastModified = time.Now()
	m.workspace.Connections = append(m.workspace.Connections, conn)
	m.workspace.LastModified = time.Now()

	logging.Info().
		Str("connection_name", conn.Name).
		Str("connection_type", string(conn.Type)).
		Msg("Connection added to workspace")

	return m.autoSave()
}

func (m *Manager) DeleteConnection(name string) error {
	idx := -1
	for i, c := range m.workspace.Connections {
		if c.Name == name {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("connection not found: %s", name)
	}

	m.workspace.Connections = append(
		m.workspace.Connections[:idx],
		m.workspace.Connections[idx+1:]...,
	)
	m.workspace.LastModified = time.Now()

	logging.Info().Str("connection_name", name).Msg("Connection deleted from workspace")

	return m.autoSave()
}

func (m *Manager) GetConnection(name string) (*models.Connection, error) {
	for i, c := range m.workspace.Connections {
		if c.Name == name {
			return &m.workspace.Connections[i], nil
		}
	}

	return nil, fmt.Errorf("connection not found: %s", name)
}

func (m *Manager) GetConnectionByName(name string) *models.Connection {
	conn, err := m.GetConnection(name)
	if err != nil {
		return nil
	}
	return conn
}

func (m *Manager) ListConnections() []models.Connection {
	connections := make([]models.Connection, len(m.workspace.Connections))
	copy(connections, m.workspace.Connections)
	return connections
}

func (m *Manager) UpdateConnection(conn models.Connection) error {
	idx := -1
	for i, c := range m.workspace.Connections {
		if c.Name == conn.Name {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("connection not found: %s", conn.Name)
	}

	conn.CreatedAt = m.workspace.Connections[idx].CreatedAt
	conn.LastModified = time.Now()
	m.workspace.Connections[idx] = conn
	m.workspace.LastModified = time.Now()

	return m.autoSave()
}

func (m *Manager) AddSavedQuery(query models.SavedQuery) error {
	if query.ID == "" {
		return errors.New("query ID is required")
	}

	if query.Name == "" {
		return errors.New("query name is required")
	}

	for _, q := range m.workspace.SavedQueries {
		if q.ID == query.ID {
			return fmt.Errorf("query with ID '%s' already exists", query.ID)
		}
	}

	query.CreatedAt = time.Now()
	query.LastModified = time.Now()
	m.workspace.SavedQueries = append(m.workspace.SavedQueries, query)
	m.workspace.LastModified = time.Now()

	return m.autoSave()
}

func (m *Manager) DeleteSavedQuery(id string) error {
	idx := -1
	for i, q := range m.workspace.SavedQueries {
		if q.ID == id {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("query not found: %s", id)
	}

	m.workspace.SavedQueries = append(
		m.workspace.SavedQueries[:idx],
		m.workspace.SavedQueries[idx+1:]...,
	)
	m.workspace.LastModified = time.Now()

	return m.autoSave()
}

func (m *Manager) GetSavedQuery(id string) (*models.SavedQuery, error) {
	for i, q := range m.workspace.SavedQueries {
		if q.ID == id {
			return &m.workspace.SavedQueries[i], nil
		}
	}

	return nil, fmt.Errorf("query not found: %s", id)
}

func (m *Manager) ListSavedQueries() []models.SavedQuery {
	queries := make([]models.SavedQuery, len(m.workspace.SavedQueries))
	copy(queries, m.workspace.SavedQueries)
	return queries
}

func (m *Manager) SearchSavedQueries(term string) []models.SavedQuery {
	var results []models.SavedQuery

	for _, q := range m.workspace.SavedQueries {
		if matchesString(q.Name, term) {
			results = append(results, q)
		}
	}

	return results
}

func (m *Manager) GetWorkspace() *models.Workspace {
	return m.workspace
}
func (m *Manager) UpdateSavedQuery(query models.SavedQuery) error {
	idx := -1
	for i, q := range m.workspace.SavedQueries {
		if q.ID == query.ID {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("query not found: %s", query.ID)
	}

	query.CreatedAt = m.workspace.SavedQueries[idx].CreatedAt
	query.LastModified = time.Now()
	m.workspace.SavedQueries[idx] = query
	m.workspace.LastModified = time.Now()

	return m.autoSave()
}

func matchesString(text, term string) bool {
	if term == "" {
		return true
	}

	textLower := strings.ToLower(text)
	termLower := strings.ToLower(term)

	for i := 0; i < len(text)-len(term)+1; i++ {
		if text[i:i+len(term)] == term || textLower[i:i+len(term)] == termLower {
			return true
		}
	}

	return false
}

func (m *Manager) GetAutocompleteEnabled() bool {
	if m.workspace.Preferences == nil {
		return true
	}
	return m.workspace.Preferences.AutocompleteEnabled
}

func (m *Manager) SetAutocompleteEnabled(enabled bool) error {
	if m.workspace.Preferences == nil {
		m.workspace.Preferences = &models.UserPreferences{}
	}

	m.workspace.Preferences.AutocompleteEnabled = enabled
	m.workspace.LastModified = time.Now()

	return m.autoSave()
}
