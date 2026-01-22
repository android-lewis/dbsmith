package components

import (
	"github.com/android-lewis/dbsmith/internal/tui/theme"
	"github.com/android-lewis/dbsmith/internal/util"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TableBuilder struct {
	table      *tview.Table
	headers    []string
	rows       [][]interface{}
	currentRow int
}

func NewTableBuilder() *TableBuilder {
	return &TableBuilder{
		table:      tview.NewTable(),
		headers:    nil,
		rows:       make([][]interface{}, 0),
		currentRow: 0,
	}
}

func (tb *TableBuilder) SetHeaders(headers ...string) *TableBuilder {
	tb.headers = headers
	return tb
}

func (tb *TableBuilder) AddRow(values ...interface{}) *TableBuilder {
	tb.rows = append(tb.rows, values)
	return tb
}

func (tb *TableBuilder) ApplyAlternatingRows() *TableBuilder {
	rowCount := tb.table.GetRowCount()
	colCount := tb.table.GetColumnCount()

	for row := 1; row < rowCount; row++ {
		var bgColor tcell.Color

		if (row-1)%2 == 0 {
			bgColor = theme.ThemeColors.Background
		} else {
			bgColor = theme.ThemeColors.BackgroundAlt
		}

		for col := 0; col < colCount; col++ {
			cell := tb.table.GetCell(row, col)
			if cell != nil {
				cell.SetBackgroundColor(bgColor)
			}
		}
	}
	return tb
}

func (tb *TableBuilder) Build() *tview.Table {
	tb.table.Clear()

	if len(tb.headers) > 0 {
		for i, header := range tb.headers {
			cell := NewHeaderCell(header)
			tb.table.SetCell(0, i, cell)
		}
		tb.currentRow = 1
	}

	for rowIdx, row := range tb.rows {
		for colIdx, val := range row {
			cellText := util.FormatValue(val)
			if cellText == "" {
				cellText = "NULL"
			}

			cell := NewDataCell(cellText)
			tb.table.SetCell(tb.currentRow+rowIdx, colIdx, cell)
		}
	}

	return tb.table
}
