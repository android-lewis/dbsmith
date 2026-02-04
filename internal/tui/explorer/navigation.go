package explorer

import (
	"github.com/android-lewis/dbsmith/internal/tui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Panel focus constants
const (
	panelSchemas = iota
	panelTables
	panelSchema
	panelIndexes
	panelData
)

// setupKeybindings configures input capture for all panels
func (e *Explorer) setupKeybindings() {
	handleAltKeys := func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			e.cyclePanel()
			return nil
		}

		// Handle 'S' key for server info (without modifiers)
		if event.Modifiers() == 0 && (event.Rune() == 's' || event.Rune() == 'S') {
			e.showServerInfo()
			return nil
		}

		if event.Modifiers()&tcell.ModAlt != 0 {
			switch event.Rune() {
			case 'h', 'H':
				e.toggleSchemas()
				return nil
			case 'i', 'I':
				e.toggleIndexes()
				return nil
			case 'd', 'D':
				e.toggleDataPreview()
				return nil
			}
		}
		return event
	}

	e.schemasList.SetInputCapture(handleAltKeys)
	e.tablesList.SetInputCapture(handleAltKeys)
	e.columnsTable.SetInputCapture(handleAltKeys)
	e.indexTable.SetInputCapture(handleAltKeys)
	e.dataTable.SetInputCapture(handleAltKeys)
}

// cyclePanel moves focus to the next visible panel
func (e *Explorer) cyclePanel() {
	panels := []int{}
	if e.showSchemas {
		panels = append(panels, panelSchemas)
	}
	panels = append(panels, panelTables)
	panels = append(panels, panelSchema)
	if e.showIndexes {
		panels = append(panels, panelIndexes)
	}
	if e.showDataPreview {
		panels = append(panels, panelData)
	}

	currentIdx := 0
	for i, p := range panels {
		if p == e.focusedPanel {
			currentIdx = i
			break
		}
	}
	e.focusedPanel = panels[(currentIdx+1)%len(panels)]
	e.updateFocus()
}

// updateFocus sets the visual focus state and application focus
func (e *Explorer) updateFocus() {
	theme.SetUnfocused(e.schemasList)
	theme.SetUnfocused(e.tablesList)
	theme.SetUnfocused(e.columnsTable)
	theme.SetUnfocused(e.indexTable)
	theme.SetUnfocused(e.dataTable)

	switch e.focusedPanel {
	case panelSchemas:
		e.setFocusedPrimitive(e.schemasList)
	case panelTables:
		e.setFocusedPrimitive(e.tablesList)
	case panelSchema:
		e.setFocusedPrimitive(e.columnsTable)
	case panelIndexes:
		e.setFocusedPrimitive(e.indexTable)
	case panelData:
		e.setFocusedPrimitive(e.dataTable)
	}
}

// setFocusedPrimitive applies focus styling and sets application focus
func (e *Explorer) setFocusedPrimitive(p tview.Primitive) {
	theme.SetFocused(p)
	e.app.SetFocus(p)
}
