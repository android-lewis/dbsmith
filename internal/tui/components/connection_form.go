package components

import (
	"fmt"

	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/rivo/tview"
)

type ConnectionFormConfig struct {
	IsEdit        bool
	ExistingConn  *models.Connection
	WorkspacePath string
}

type ConnectionFormManager struct {
	pages *tview.Pages
	app   *tview.Application

	onSubmit  func(conn models.Connection, password string, isEdit bool) error
	onSuccess func(message string)
}

func NewConnectionFormManager(pages *tview.Pages, app *tview.Application) *ConnectionFormManager {
	return &ConnectionFormManager{
		pages: pages,
		app:   app,
	}
}

func (m *ConnectionFormManager) SetCallbacks(
	onSubmit func(conn models.Connection, password string, isEdit bool) error,
	onSuccess func(message string),
) {
	m.onSubmit = onSubmit
	m.onSuccess = onSuccess
}

func (m *ConnectionFormManager) ShowNewConnectionForm() {
	m.showConnectionForm(ConnectionFormConfig{
		IsEdit: false,
	})
}

func (m *ConnectionFormManager) ShowEditConnectionForm(conn *models.Connection) {
	if conn == nil {
		ShowError(m.pages, m.app, fmt.Errorf("no connection provided for editing"))
		return
	}
	m.showConnectionForm(ConnectionFormConfig{
		IsEdit:       true,
		ExistingConn: conn,
	})
}

func (m *ConnectionFormManager) showConnectionForm(config ConnectionFormConfig) {
	title := m.getFormTitle(config.IsEdit)
	fields := m.buildFormFields(config)

	dialog := NewFormDialog(m.pages, m.app, FormDialogConfig{
		Title:       title,
		Fields:      fields,
		SubmitLabel: "Save",
		CancelLabel: "Cancel",
		PageName:    "connection-form",
		ModalWidth:  60,
		OnSubmit:    m.createSubmitHandler(config.IsEdit),
	})

	dialog.Show()
}

func (m *ConnectionFormManager) getFormTitle(isEdit bool) string {
	if isEdit {
		return " Edit Connection "
	}
	return " New Connection "
}

func (m *ConnectionFormManager) buildFormFields(config ConnectionFormConfig) []FormField {
	defaults := m.getDefaultValues(config)
	dbTypes := []string{"postgres", "mysql", "sqlite"}
	sslModes := []string{"disable", "prefer", "require"}

	return []FormField{
		{Type: FieldTypeInput, Label: "Name", InitialValue: defaults["name"], FieldWidth: 30},
		{
			Type:         FieldTypeDropDown,
			Label:        "Type",
			Options:      dbTypes,
			InitialIndex: findIndex(dbTypes, defaults["dbType"]),
		},
		{Type: FieldTypeInput, Label: "Host", InitialValue: defaults["host"], FieldWidth: 30},
		{Type: FieldTypeInput, Label: "Port", InitialValue: defaults["port"], FieldWidth: 10},
		{Type: FieldTypeInput, Label: "Database", InitialValue: defaults["database"], FieldWidth: 30},
		{Type: FieldTypeInput, Label: "Username", InitialValue: defaults["username"], FieldWidth: 30},
		{Type: FieldTypePassword, Label: "Password", FieldWidth: 30},
		{
			Type:         FieldTypeDropDown,
			Label:        "SSL",
			Options:      sslModes,
			InitialIndex: findIndex(sslModes, defaults["ssl"]),
		},
	}
}

func (m *ConnectionFormManager) getDefaultValues(config ConnectionFormConfig) map[string]string {
	defaults := map[string]string{
		"name":     "",
		"dbType":   "postgres",
		"host":     "localhost",
		"port":     "5432",
		"database": "",
		"username": "",
		"ssl":      "prefer",
	}

	if config.IsEdit && config.ExistingConn != nil {
		conn := config.ExistingConn
		defaults["name"] = conn.Name
		defaults["dbType"] = string(conn.Type)
		defaults["host"] = conn.Host
		defaults["port"] = fmt.Sprintf("%d", conn.Port)
		defaults["database"] = conn.Database
		defaults["username"] = conn.Username
		defaults["ssl"] = conn.SSL
	}

	return defaults
}

func (m *ConnectionFormManager) createSubmitHandler(isEdit bool) func(map[string]string) error {
	return func(values map[string]string) error {
		if err := m.validateFormValues(values, isEdit); err != nil {
			return err
		}

		conn := m.buildConnectionFromValues(values)

		if m.onSubmit != nil {
			if err := m.onSubmit(conn, values["Password"], isEdit); err != nil {
				return err
			}
		}

		if m.onSuccess != nil {
			action := "created"
			if isEdit {
				action = "updated"
			}
			m.onSuccess(fmt.Sprintf("Connection %s: %s", action, conn.Name))
		}

		return nil
	}
}

func (m *ConnectionFormManager) validateFormValues(values map[string]string, isEdit bool) error {
	if values["Name"] == "" {
		return fmt.Errorf("connection name is required")
	}
	if values["Host"] == "" {
		return fmt.Errorf("host is required")
	}
	if values["Database"] == "" {
		return fmt.Errorf("database name is required")
	}
	if !isEdit && values["Password"] == "" {
		return fmt.Errorf("password is required for new connections")
	}
	return nil
}

func (m *ConnectionFormManager) buildConnectionFromValues(values map[string]string) models.Connection {
	port := 5432
	_, _ = fmt.Sscanf(values["Port"], "%d", &port)

	secretKeyID := fmt.Sprintf("dbsmith_%s_%s", values["Name"], values["Type"])

	return models.Connection{
		Name:        values["Name"],
		Type:        models.ConnectionType(values["Type"]),
		Host:        values["Host"],
		Port:        port,
		Database:    values["Database"],
		Username:    values["Username"],
		SecretKeyID: secretKeyID,
		SSL:         values["SSL"],
	}
}

func findIndex(slice []string, value string) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return 0
}
