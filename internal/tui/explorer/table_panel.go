package explorer

import (
	"context"
	"fmt"

	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/tui/components"
	"github.com/android-lewis/dbsmith/internal/tui/constants"
	"github.com/android-lewis/dbsmith/internal/tui/theme"
	"github.com/rivo/tview"
)

// buildTablesList creates and configures the tables list widget
func (e *Explorer) buildTablesList() *tview.List {
	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)

	list.SetBorder(true).
		SetTitle(" Tables ").
		SetTitleAlign(tview.AlignLeft)

	list.SetSelectedBackgroundColor(theme.ThemeColors.Selection).
		SetSelectedTextColor(theme.ThemeColors.SelectionText)

	return list
}

// loadTablesForSchema fetches and displays tables for the given schema
func (e *Explorer) loadTablesForSchema(schemaName string) {
	e.selectedSchema = schemaName
	e.tablesList.Clear()

	if e.dbApp.Explorer == nil {
		e.tablesList.AddItem("No connection", "", 0, nil)
		return
	}

	e.tablesList.AddItem("Loading...", "", 0, nil)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutSchemaLoad)
		defer cancel()

		schema := models.Schema{Name: schemaName}
		tables, err := e.dbApp.Explorer.GetTables(ctx, schema)
		if err != nil {
			e.app.QueueUpdateDraw(func() {
				e.tablesList.Clear()
				e.tablesList.AddItem("Error loading tables", "", 0, nil)
				components.ShowError(e.pages, e.app, fmt.Errorf("failed to load tables: %w", err))
			})
			return
		}

		e.app.QueueUpdateDraw(func() {
			e.tablesList.Clear()

			if len(tables) == 0 {
				e.tablesList.AddItem("No tables", "", 0, nil)
				return
			}

			for _, table := range tables {
				localTableName := table.Name
				e.tablesList.AddItem(localTableName, "", 0, func() {
					e.loadTableDetails(localTableName)
				})
			}

			if len(tables) > 0 {
				e.tablesList.SetCurrentItem(0)
				e.loadTableDetails(tables[0].Name)
			}
		})
	}()
}

// loadTableDetails loads columns, indexes, and data preview for a table
func (e *Explorer) loadTableDetails(tableName string) {
	e.selectedTable = tableName
	e.dataOffset = 0

	go e.loadColumns(tableName)

	if e.showDataPreview {
		go e.loadDataPreview(tableName)
	}

	if e.showIndexes {
		go e.loadIndexes(tableName)
	}
}
