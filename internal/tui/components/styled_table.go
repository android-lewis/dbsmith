package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/android-lewis/dbsmith/internal/tui/theme"
)

func NewHeaderCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetTextColor(theme.ThemeColors.Primary).
		SetAttributes(tcell.AttrBold).
		SetSelectable(false)
}

func NewDataCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetTextColor(theme.ThemeColors.Foreground).
		SetExpansion(1)
}
