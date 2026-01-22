package components

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/tui/constants"
	"github.com/android-lewis/dbsmith/internal/tui/theme"
)

type QueryResultContent struct {
	result       *models.QueryResult
	cache        *LRUCellCache
	maxCellLen   int
	altRowColors bool
}

func NewQueryResultContent(result *models.QueryResult) *QueryResultContent {
	return NewQueryResultContentWithMaxLen(result, constants.MaxCellDisplayLen)
}

func NewQueryResultContentWithMaxLen(result *models.QueryResult, maxCellLen int) *QueryResultContent {
	return &QueryResultContent{
		result:     result,
		cache:      NewLRUCellCache(constants.DefaultCellCacheSize),
		maxCellLen: maxCellLen,
	}
}

func (q *QueryResultContent) ApplyAlternatingRowColors() {
	q.altRowColors = true
}

func (q *QueryResultContent) GetCell(row, col int) *tview.TableCell {
	if cell := q.cache.Get(row, col); cell != nil {
		return cell
	}

	var cell *tview.TableCell

	if row == 0 {
		if col < len(q.result.Columns) {
			cell = NewHeaderCell(q.result.Columns[col]).SetExpansion(1)
		} else {
			cell = tview.NewTableCell("")
		}
	} else {
		dataRow := row - 1
		if dataRow < len(q.result.Rows) && col < len(q.result.Rows[dataRow]) {
			cellText := q.formatCellValue(q.result.Rows[dataRow][col])
			cell = NewDataCell(cellText)

			if q.altRowColors {
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

	q.cache.Set(row, col, cell)
	return cell
}

func (q *QueryResultContent) GetRowCount() int {
	return len(q.result.Rows) + 1
}

func (q *QueryResultContent) GetColumnCount() int {
	return len(q.result.Columns)
}

func (q *QueryResultContent) SetCell(row, col int, cell *tview.TableCell) {
	q.cache.Set(row, col, cell)
}

func (q *QueryResultContent) RemoveRow(row int) {}

func (q *QueryResultContent) RemoveColumn(col int) {}

func (q *QueryResultContent) InsertRow(row int) {}

func (q *QueryResultContent) InsertColumn(col int) {}

func (q *QueryResultContent) Clear() {
	q.cache.Clear()
}

func (q *QueryResultContent) formatCellValue(val interface{}) string {
	var cellText string

	switch v := val.(type) {
	case nil:
		return "NULL"
	case []uint8:
		cellText = string(v)
	case float32:
		cellText = strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		cellText = strconv.FormatFloat(v, 'f', -1, 64)
	default:
		cellText = fmt.Sprintf("%v", val)
		if cellText == "" || cellText == "<nil>" {
			return "NULL"
		}
	}

	if len(cellText) > q.maxCellLen {
		cellText = cellText[:q.maxCellLen-3] + "..."
	}

	return cellText
}
