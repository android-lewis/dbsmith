package explorer

import (
	"context"
	"fmt"

	"github.com/android-lewis/dbsmith/internal/tui/components"
	"github.com/android-lewis/dbsmith/internal/tui/constants"
	"github.com/android-lewis/dbsmith/internal/tui/theme"
	"github.com/rivo/tview"
)

// buildSchemasList creates and configures the schemas list widget
func (e *Explorer) buildSchemasList() *tview.List {
	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)

	list.SetBorder(true).
		SetTitle(" Schemas ").
		SetTitleAlign(tview.AlignLeft)

	list.SetSelectedBackgroundColor(theme.ThemeColors.Selection).
		SetSelectedTextColor(theme.ThemeColors.SelectionText)

	return list
}

// loadSchemas fetches and displays available database schemas
func (e *Explorer) loadSchemas() {
	e.schemasList.Clear()

	if e.dbApp.Explorer == nil {
		e.schemasList.AddItem("No connection", "", 0, nil)
		return
	}

	e.schemasList.AddItem("Loading...", "", 0, nil)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutSchemaLoad)
		defer cancel()

		schemas, err := e.dbApp.Explorer.GetSchemas(ctx)
		if err != nil {
			e.app.QueueUpdateDraw(func() {
				e.schemasList.Clear()
				e.schemasList.AddItem("Error loading schemas", "", 0, nil)
				components.ShowError(e.pages, e.app, fmt.Errorf("failed to load schemas: %w", err))
			})
			return
		}

		e.app.QueueUpdateDraw(func() {
			e.schemasList.Clear()

			if len(schemas) == 0 {
				e.schemasList.AddItem("No schemas", "", 0, nil)
				return
			}

			for _, schema := range schemas {
				localSchemaName := schema.Name
				e.schemasList.AddItem(localSchemaName, "", 0, func() {
					e.loadTablesForSchema(localSchemaName)
				})
			}

			if len(schemas) > 0 {
				e.schemasList.SetCurrentItem(0)
				e.selectedSchema = schemas[0].Name
				e.loadTablesForSchema(schemas[0].Name)
			}
		})
	}()
}

// toggleSchemas shows or hides the schemas panel
func (e *Explorer) toggleSchemas() {
	e.showSchemas = !e.showSchemas

	if !e.showSchemas && e.focusedPanel == panelSchemas {
		e.focusedPanel = panelTables
	}

	e.leftFlex.Clear()
	if e.showSchemas {
		e.leftFlex.
			AddItem(e.schemasList, 0, 1, false).
			AddItem(e.tablesList, 0, 2, false)

		if len(e.schemasList.FindItems("", "", false, false)) == 0 {
			go e.loadSchemas()
		}
	} else {
		e.leftFlex.AddItem(e.tablesList, 0, 1, false)
	}

	e.updateFocus()
}
