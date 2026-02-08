package explorer

import (
	"context"
	"fmt"

	"github.com/android-lewis/dbsmith/internal/tui/components"
	"github.com/android-lewis/dbsmith/internal/tui/constants"
	"github.com/android-lewis/dbsmith/internal/tui/theme"
	"github.com/android-lewis/dbsmith/internal/tui/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// buildDataTable creates and configures the data preview table widget
func (e *Explorer) buildDataTable() *tview.Table {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)

	table.SetBorder(true).
		SetTitle(" Data Preview ").
		SetTitleAlign(tview.AlignLeft)

	table.SetSelectedStyle(tcell.StyleDefault.
		Background(theme.ThemeColors.Selection).
		Foreground(theme.ThemeColors.SelectionText))

	return table
}

// loadDataPreview fetches and displays a preview of table data
func (e *Explorer) loadDataPreview(tableName string) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutSchemaLoad)
	defer cancel()

	result, err := e.dbApp.Explorer.GetTableData(ctx, tableName, e.dataLimit, e.dataOffset)
	if err != nil {
		e.app.QueueUpdateDraw(func() {
			e.dataTable.Clear()
			e.dataTable.SetContent(nil)
			e.dataTable.SetCell(0, 0, tview.NewTableCell(fmt.Sprintf("Error: %v", err)))
		})
		return
	}

	e.app.QueueUpdateDraw(func() {
		e.dataTable.Clear()

		startRow := e.dataOffset + 1
		endRow := e.dataOffset + len(result.Rows)
		title := fmt.Sprintf(" Data Preview [%s-%s of %s] ",
			utils.FormatNumber(int64(startRow)),
			utils.FormatNumber(int64(endRow)),
			utils.FormatNumber(result.RowCount))
		e.dataTable.SetTitle(title)

		maxCellWidth := e.dbApp.Config.UI.MaxPreviewCellWidth
		if maxCellWidth <= 0 {
			maxCellWidth = constants.MaxPreviewCellLen
		}
		content := components.NewQueryResultContentWithMaxLen(result, maxCellWidth)
		content.ApplyAlternatingRowColors()
		e.dataTable.SetContent(content)
	})
}

// toggleDataPreview shows or hides the data preview panel
func (e *Explorer) toggleDataPreview() {
	e.showDataPreview = !e.showDataPreview

	if !e.showDataPreview && e.focusedPanel == panelData {
		e.focusedPanel = panelTables
	}

	if e.showDataPreview && e.selectedTable != "" {
		go e.loadDataPreview(e.selectedTable)
	}

	e.rebuildLayout()
}

// rebuildLayout reconstructs the main layout based on visibility settings
func (e *Explorer) rebuildLayout() {
	e.mainFlex.Clear()

	if e.showDataPreview {
		e.mainFlex.
			AddItem(e.leftFlex, 25, 0, false).
			AddItem(e.schemaFlex, 35, 0, false).
			AddItem(e.dataTable, 0, 1, false)
	} else {
		e.mainFlex.
			AddItem(e.leftFlex, 25, 0, false).
			AddItem(e.schemaFlex, 0, 1, false)
	}

	e.updateFocus()
}
