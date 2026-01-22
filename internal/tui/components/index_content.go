package components

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/tui/constants"
	"github.com/android-lewis/dbsmith/internal/tui/theme"
)

type IndexContent struct {
	indexes      []models.Index
	cache        *LRUCellCache
	altRowColors bool
}

var indexContentHeaders = []string{"Name", "Columns", "Unique", "Type"}

func NewIndexContent(indexes []models.Index) *IndexContent {
	return &IndexContent{
		indexes: indexes,
		cache:   NewLRUCellCache(constants.DefaultCellCacheSize),
	}
}

func (i *IndexContent) ApplyAlternatingRowColors() {
	i.altRowColors = true
}

func (i *IndexContent) GetCell(row, col int) *tview.TableCell {

	if cell := i.cache.Get(row, col); cell != nil {
		return cell
	}

	var cell *tview.TableCell

	if row == 0 {
		if col < len(indexContentHeaders) {
			cell = NewHeaderCell(indexContentHeaders[col]).SetExpansion(1)
		} else {
			cell = tview.NewTableCell("")
		}
	} else {

		dataRow := row - 1
		if dataRow < len(i.indexes) {
			cell = i.createIndexCell(dataRow, col)

			if i.altRowColors {
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

	i.cache.Set(row, col, cell)
	return cell
}

func (i *IndexContent) createIndexCell(dataRow, col int) *tview.TableCell {
	index := i.indexes[dataRow]

	switch col {
	case 0:
		return NewDataCell(index.Name)
	case 1:
		return NewDataCell(strings.Join(index.Columns, ", "))

	case 2:
		cell := NewDataCell("")
		if index.IsUnique {
			cell.SetText(theme.Icons.Check)
			cell.SetTextColor(theme.ThemeColors.Success)
		}
		return cell

	case 3:
		return NewDataCell(index.Type)

	default:
		return tview.NewTableCell("")
	}
}

func (i *IndexContent) GetRowCount() int {
	return len(i.indexes) + 1
}

func (i *IndexContent) GetColumnCount() int {
	return len(indexContentHeaders)
}

func (i *IndexContent) SetCell(row, col int, cell *tview.TableCell) {
	i.cache.Set(row, col, cell)
}

func (i *IndexContent) RemoveRow(row int) {}

func (i *IndexContent) RemoveColumn(col int) {}

func (i *IndexContent) InsertRow(row int) {}

func (i *IndexContent) InsertColumn(col int) {}

func (i *IndexContent) Clear() {
	i.cache.Clear()
}
