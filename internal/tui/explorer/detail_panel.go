package explorer

import (
	"context"

	"github.com/android-lewis/dbsmith/internal/tui/components"
	"github.com/android-lewis/dbsmith/internal/tui/constants"
	"github.com/android-lewis/dbsmith/internal/tui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// buildColumnsTable creates and configures the columns table widget
func (e *Explorer) buildColumnsTable() *tview.Table {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(false, false)

	table.SetBorder(true).
		SetTitle(" Columns ").
		SetTitleAlign(tview.AlignLeft)

	return table
}

// buildIndexTable creates and configures the index table widget
func (e *Explorer) buildIndexTable() *tview.Table {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	table.SetBorder(true).
		SetTitle(" Indexes ").
		SetTitleAlign(tview.AlignLeft)

	table.SetSelectedStyle(tcell.StyleDefault.
		Background(theme.ThemeColors.Selection).
		Foreground(theme.ThemeColors.SelectionText))

	return table
}

// loadColumns fetches and displays column information for a table
func (e *Explorer) loadColumns(tableName string) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutSchemaLoad)
	defer cancel()

	schema, err := e.dbApp.Explorer.GetTableColumns(ctx, tableName)
	if err != nil {
		e.app.QueueUpdateDraw(func() {
			e.columnsTable.Clear()
			e.columnsTable.SetContent(nil)
			e.columnsTable.SetCell(0, 0, tview.NewTableCell("Error loading columns"))
		})
		return
	}

	e.app.QueueUpdateDraw(func() {
		e.columnsTable.Clear()

		content := components.NewColumnContent(schema.Columns)
		content.ApplyAlternatingRowColors()
		e.columnsTable.SetContent(content)
	})
}

// loadIndexes fetches and displays index information for a table
func (e *Explorer) loadIndexes(tableName string) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutSchemaLoad)
	defer cancel()

	indexes, err := e.dbApp.Explorer.GetTableIndexes(ctx, tableName)
	if err != nil {
		e.app.QueueUpdateDraw(func() {
			e.indexTable.Clear()
			e.indexTable.SetContent(nil)
			e.indexTable.SetCell(0, 0, tview.NewTableCell("Error loading indexes"))
		})
		return
	}

	e.app.QueueUpdateDraw(func() {
		e.indexTable.Clear()

		if len(indexes) == 0 {
			e.indexTable.SetContent(nil)
			e.indexTable.SetCell(0, 0, tview.NewTableCell("No indexes"))
			return
		}

		content := components.NewIndexContent(indexes)
		content.ApplyAlternatingRowColors()
		e.indexTable.SetContent(content)
	})
}

// toggleIndexes shows or hides the indexes panel
func (e *Explorer) toggleIndexes() {
	e.showIndexes = !e.showIndexes

	if !e.showIndexes && e.focusedPanel == panelIndexes {
		e.focusedPanel = panelSchema
	}

	e.schemaFlex.Clear()
	if e.showIndexes {
		e.schemaFlex.
			AddItem(e.columnsTable, 0, 2, false).
			AddItem(e.indexTable, 0, 1, false)

		if e.selectedTable != "" {
			go e.loadIndexes(e.selectedTable)
		}
	} else {
		e.schemaFlex.AddItem(e.columnsTable, 0, 1, false)
	}

	e.updateFocus()
}
