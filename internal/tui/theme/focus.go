package theme

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type FocusManager struct {
	focusedPrimitive tview.Primitive
}

func NewFocusManager() *FocusManager {
	return &FocusManager{}
}

func SetFocused(primitive tview.Primitive) {
	if primitive == nil {
		return
	}

	switch p := primitive.(type) {
	case *tview.Box:
		applyFocusedBox(p)
	case *tview.TextView:
		applyFocusedTextView(p)
	case *tview.Table:
		applyFocusedTable(p)
	case *tview.List:
		applyFocusedList(p)
	case *tview.InputField:
		applyFocusedInputField(p)
	case *tview.TextArea:
		applyFocusedTextArea(p)
	case *tview.Flex:
		if box := p.Box; box != nil {
			applyFocusedBox(box)
		}
	}
}

func SetUnfocused(primitive tview.Primitive) {
	if primitive == nil {
		return
	}

	switch p := primitive.(type) {
	case *tview.Box:
		applyUnfocusedBox(p)
	case *tview.TextView:
		applyUnfocusedTextView(p)
	case *tview.Table:
		applyUnfocusedTable(p)
	case *tview.List:
		applyUnfocusedList(p)
	case *tview.InputField:
		applyUnfocusedInputField(p)
	case *tview.TextArea:
		applyUnfocusedTextArea(p)
	case *tview.Flex:
		if box := p.Box; box != nil {
			applyUnfocusedBox(box)
		}
	}
}

func TransitionFocus(from, to tview.Primitive) {
	if from != nil {
		SetUnfocused(from)
	}
	if to != nil {
		SetFocused(to)
	}
}

func applyFocusedBox(box *tview.Box) {
	box.SetBorderColor(ThemeColors.BorderFocus)
	box.SetTitleColor(ThemeColors.Primary)
	box.SetBackgroundColor(ThemeColors.BackgroundAlt)
}

func applyUnfocusedBox(box *tview.Box) {
	box.SetBorderColor(ThemeColors.Border)
	box.SetTitleColor(ThemeColors.ForegroundMuted)
	box.SetBackgroundColor(ThemeColors.Background)
}

func applyFocusedTextView(tv *tview.TextView) {
	applyFocusedBox(tv.Box)
	tv.SetTextColor(ThemeColors.Foreground)
}

func applyUnfocusedTextView(tv *tview.TextView) {
	applyUnfocusedBox(tv.Box)
	tv.SetTextColor(ThemeColors.Foreground)
}

func applyFocusedTable(table *tview.Table) {
	applyFocusedBox(table.Box)
	table.SetSelectedStyle(tcell.StyleDefault.
		Foreground(ThemeColors.SelectionText).
		Background(ThemeColors.Primary).
		Bold(true))
}

func applyUnfocusedTable(table *tview.Table) {
	applyUnfocusedBox(table.Box)
	table.SetSelectedStyle(tcell.StyleDefault.
		Foreground(ThemeColors.SelectionText).
		Background(ThemeColors.Selection))
}

func applyFocusedList(list *tview.List) {
	applyFocusedBox(list.Box)
	list.SetSelectedBackgroundColor(ThemeColors.Primary)
	list.SetSelectedTextColor(ThemeColors.SelectionText)
}

func applyUnfocusedList(list *tview.List) {
	applyUnfocusedBox(list.Box)
	list.SetSelectedBackgroundColor(ThemeColors.Selection)
	list.SetSelectedTextColor(ThemeColors.SelectionText)
}

func applyFocusedInputField(input *tview.InputField) {
	input.SetFieldBackgroundColor(ThemeColors.BackgroundAlt)
	input.SetFieldTextColor(ThemeColors.Foreground)
	input.SetLabelColor(ThemeColors.Primary)
}

func applyUnfocusedInputField(input *tview.InputField) {
	input.SetFieldBackgroundColor(ThemeColors.Background)
	input.SetFieldTextColor(ThemeColors.Foreground)
	input.SetLabelColor(ThemeColors.ForegroundMuted)
}

func applyFocusedTextArea(ta *tview.TextArea) {
	applyFocusedBox(ta.Box)
}

func applyUnfocusedTextArea(ta *tview.TextArea) {
	applyUnfocusedBox(ta.Box)
}

func (fm *FocusManager) SetCurrentFocus(primitive tview.Primitive) {
	if fm.focusedPrimitive == primitive {
		return
	}

	if fm.focusedPrimitive != nil {
		SetUnfocused(fm.focusedPrimitive)
	}

	fm.focusedPrimitive = primitive
	if fm.focusedPrimitive != nil {
		SetFocused(fm.focusedPrimitive)
	}
}

func (fm *FocusManager) GetCurrentFocus() tview.Primitive {
	return fm.focusedPrimitive
}

func (fm *FocusManager) ClearFocus() {
	if fm.focusedPrimitive != nil {
		SetUnfocused(fm.focusedPrimitive)
		fm.focusedPrimitive = nil
	}
}
