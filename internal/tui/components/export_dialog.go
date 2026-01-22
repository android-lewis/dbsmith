package components

import (
	"fmt"
	"os"

	"github.com/android-lewis/dbsmith/internal/exporter"
	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/rivo/tview"
)

type ExportManager struct {
	pages *tview.Pages
	app   *tview.Application

	getLastResult func() *models.QueryResult
	getDialect    func() string
}

func NewExportManager(pages *tview.Pages, app *tview.Application) *ExportManager {
	return &ExportManager{
		pages: pages,
		app:   app,
	}
}

func (m *ExportManager) SetCallbacks(
	getLastResult func() *models.QueryResult,
	getDialect func() string,
) {
	m.getLastResult = getLastResult
	m.getDialect = getDialect
}

func (m *ExportManager) ShowExportDialog() {
	if m.getLastResult == nil {
		ShowError(m.pages, m.app, fmt.Errorf("no results to export"))
		return
	}

	lastResult := m.getLastResult()
	if lastResult == nil {
		ShowError(m.pages, m.app, fmt.Errorf("no results to export"))
		return
	}

	fields := []FormField{
		{
			Type:         FieldTypeDropDown,
			Label:        "Format",
			Options:      []string{"csv", "json"},
			InitialIndex: 0,
		},
		{Type: FieldTypeInput, Label: "File Path", FieldWidth: 40},
	}

	dialog := NewFormDialog(m.pages, m.app, FormDialogConfig{
		Title:         " Export Results ",
		Fields:        fields,
		SubmitLabel:   "Export",
		CancelLabel:   "Cancel",
		PageName:      "export-dialog",
		ModalWidth:    60,
		ModalHeight:   11,
		EscapeToClose: true,
		OnSubmit: func(values map[string]string) error {
			filePath := values["File Path"]
			if filePath == "" {
				return fmt.Errorf("file path is required")
			}

			selectedFormat := values["Format"]

			if err := m.exportResults(lastResult, filePath, selectedFormat); err != nil {
				return err
			}

			ShowInfo(m.pages, m.app, fmt.Sprintf("Results exported to %s as %s", filePath, selectedFormat))

			return nil
		},
	})

	form := dialog.GetForm()

	if form.GetFormItemCount() > 0 {
		formatDropdown := form.GetFormItem(0).(*tview.DropDown)
		formatDropdown.SetSelectedFunc(func(option string, index int) {
			for form.GetFormItemCount() > 2 {
				form.RemoveFormItem(2)
			}
		})
	}

	dialog.Show()
}

func (m *ExportManager) exportResults(result *models.QueryResult, filePath, format string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	return exporter.ExportToFormat(file, result, format)
}
