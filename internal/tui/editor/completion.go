package editor

import (
	"fmt"
	"strings"

	"github.com/android-lewis/dbsmith/internal/autocomplete"
	"github.com/android-lewis/dbsmith/internal/tui/theme"
	"github.com/rivo/tview"
)

const completionPageName = "editor_completion_popup"

type completionOverlay struct {
	pages      *tview.Pages
	view       *tview.TextView
	grid       *tview.Grid
	items      []autocomplete.Item
	selected   int
	visible    bool
	maxVisible int
}

func newCompletionOverlay(pages *tview.Pages) *completionOverlay {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	view.SetBorder(true).
		SetTitle(" Completions ")
	view.SetBackgroundColor(theme.ThemeColors.BackgroundAlt)
	view.SetTextColor(theme.ThemeColors.Foreground)

	grid := tview.NewGrid()

	return &completionOverlay{
		pages:      pages,
		view:       view,
		grid:       grid,
		maxVisible: 8,
	}
}

func (c *completionOverlay) Show(items []autocomplete.Item, selected int) {
	if len(items) == 0 {
		c.Hide()
		return
	}

	if selected < 0 {
		selected = 0
	}
	if selected >= len(items) {
		selected = len(items) - 1
	}

	c.items = items
	c.selected = selected
	c.visible = true

	c.render()
	c.pages.RemovePage(completionPageName)
	c.pages.AddPage(completionPageName, c.grid, true, true)
}

func (c *completionOverlay) Hide() {
	if !c.visible {
		return
	}
	c.visible = false
	c.pages.RemovePage(completionPageName)
}

func (c *completionOverlay) IsVisible() bool {
	return c.visible
}

func (c *completionOverlay) SetSelected(index int) {
	if !c.visible || len(c.items) == 0 {
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(c.items) {
		index = len(c.items) - 1
	}
	c.selected = index
	c.render()
}

func (c *completionOverlay) render() {
	if !c.visible {
		return
	}

	start := 0
	if c.selected >= c.maxVisible {
		start = c.selected - c.maxVisible + 1
	}
	end := start + c.maxVisible
	if end > len(c.items) {
		end = len(c.items)
	}

	lines := make([]string, 0, end-start)
	width := 0

	for idx := start; idx < end; idx++ {
		item := c.items[idx]
		line := formatCompletionLine(item)
		if idx == c.selected {
			line = highlightCompletionLine(line)
		}

		lines = append(lines, line)
		lineLen := visibleLength(line)
		if lineLen > width {
			width = lineLen
		}
	}

	c.view.SetText(strings.Join(lines, "\n"))

	height := len(lines) + 2
	if height < 3 {
		height = 3
	}
	width += 4
	if width < 24 {
		width = 24
	}

	c.grid.Clear()
	c.grid.SetRows(0, height, 0)
	c.grid.SetColumns(0, width, 0)
	c.grid.AddItem(c.view, 1, 1, 1, 1, 0, 0, false)
}

func formatCompletionLine(item autocomplete.Item) string {
	if item.Detail == "" {
		return item.Label
	}
	return fmt.Sprintf("%s [#%06x]%s[-]", item.Label, theme.ThemeColors.ForegroundMuted.Hex(), item.Detail)
}

func highlightCompletionLine(text string) string {
	return fmt.Sprintf("[#%06x:#%06x]%s[-:-]",
		theme.ThemeColors.SelectionText.Hex(),
		theme.ThemeColors.Selection.Hex(),
		text)
}

func visibleLength(text string) int {
	inTag := false
	length := 0
	for _, r := range text {
		switch r {
		case '[':
			inTag = true
		case ']':
			inTag = false
		default:
			if !inTag {
				length++
			}
		}
	}
	return length
}
