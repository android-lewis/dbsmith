package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/tui/constants"
	"github.com/android-lewis/dbsmith/internal/tui/theme"
)

type ColumnContent struct {
	columns      []models.Column
	cache        *LRUCellCache
	altRowColors bool
}

var columnContentHeaders = []string{"Column", "Type", "Nullable", "Default"}

func NewColumnContent(columns []models.Column) *ColumnContent {
	return &ColumnContent{
		columns: columns,
		cache:   NewLRUCellCache(constants.DefaultCellCacheSize),
	}
}

func (c *ColumnContent) ApplyAlternatingRowColors() {
	c.altRowColors = true
}

func (c *ColumnContent) GetCell(row, col int) *tview.TableCell {

	if cell := c.cache.Get(row, col); cell != nil {
		return cell
	}

	var cell *tview.TableCell

	if row == 0 {

		if col < len(columnContentHeaders) {
			cell = NewHeaderCell(columnContentHeaders[col]).SetExpansion(1)
		} else {
			cell = tview.NewTableCell("")
		}
	} else {

		dataRow := row - 1
		if dataRow < len(c.columns) {
			cell = c.createColumnCell(dataRow, col)

			if c.altRowColors {
				var bgColor tcell.Color
				if dataRow%2 == 0 {
					bgColor = theme.ThemeColors.Background
				} else {
					bgColor = theme.ThemeColors.BackgroundAlt
				}
				cell.SetBackgroundColor(bgColor)
			}
		} else {
			cell = tview.NewTableCell("")
		}
	}

	c.cache.Set(row, col, cell)
	return cell
}

func (c *ColumnContent) createColumnCell(dataRow, col int) *tview.TableCell {
	column := c.columns[dataRow]

	switch col {
	case 0:
		return NewDataCell(column.Name)
	case 1:
		return NewDataCell(column.Type)

	case 2:
		nullable := "NO"
		if column.Nullable {
			nullable = "YES"
		}
		return NewDataCell(nullable)

	case 3:
		defaultVal := column.Default
		if defaultVal == "" {
			defaultVal = "-"
		}
		return NewDataCell(defaultVal)

	default:
		return tview.NewTableCell("")
	}
}

func (c *ColumnContent) GetRowCount() int {
	return len(c.columns) + 1
}

func (c *ColumnContent) GetColumnCount() int {
	return len(columnContentHeaders)
}

func (c *ColumnContent) SetCell(row, col int, cell *tview.TableCell) {
	c.cache.Set(row, col, cell)
}

func (c *ColumnContent) RemoveRow(row int) {}

func (c *ColumnContent) RemoveColumn(col int) {}

func (c *ColumnContent) InsertRow(row int) {}

func (c *ColumnContent) InsertColumn(col int) {}

func (c *ColumnContent) Clear() {
	c.cache.Clear()
}
