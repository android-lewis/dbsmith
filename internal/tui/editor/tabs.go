package editor

import (
	"fmt"
	"strings"

	"github.com/android-lewis/dbsmith/internal/app"
	"github.com/android-lewis/dbsmith/internal/tui/components"
	"github.com/android-lewis/dbsmith/internal/tui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TabState struct {
	id            int
	name          string
	editor        *Editor
	modified      bool
	running       bool
	savedQueryID  string
	savedQuerySQL string
}

type EditorTabs struct {
	app       *tview.Application
	pages     *tview.Pages
	dbApp     *app.App
	helpBar   *components.HelpBar
	statusBar *components.StatusBar

	tabs        []*TabState
	activeTab   int
	nextTabID   int
	mainFlex    *tview.Flex
	tabBar      *tview.TextView
	contentFlex *tview.Flex
}

func NewEditorTabs(app *tview.Application, pages *tview.Pages, dbApp *app.App, helpBar *components.HelpBar, statusBar *components.StatusBar) *EditorTabs {
	et := &EditorTabs{
		app:       app,
		pages:     pages,
		dbApp:     dbApp,
		helpBar:   helpBar,
		statusBar: statusBar,
		tabs:      make([]*TabState, 0),
		activeTab: 0,
		nextTabID: 1,
	}

	et.buildUI()
	et.createNewTab("New Query")

	return et
}

func (et *EditorTabs) buildUI() {
	et.tabBar = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(false)
	et.tabBar.SetBorder(false)

	et.contentFlex = tview.NewFlex().
		SetDirection(tview.FlexRow)

	et.mainFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(et.tabBar, 1, 0, false).
		AddItem(et.contentFlex, 0, 1, true)

	et.setupGlobalKeybindings()
}

func (et *EditorTabs) setupGlobalKeybindings() {
	et.mainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Modifiers()&tcell.ModAlt != 0 {
			switch event.Rune() {
			case 't', 'T':
				et.createNewTab(fmt.Sprintf("Query %d", et.nextTabID))
				return nil
			case 'w', 'W':
				et.closeCurrentTab()
				return nil
			case 'r', 'R':
				et.renameCurrentTab()
				return nil
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				tabIndex := int(event.Rune() - '1')
				if tabIndex < len(et.tabs) {
					et.switchToTab(tabIndex)
				}
				return nil
			}
		}

		if event.Key() == tcell.KeyLeft && event.Modifiers()&tcell.ModAlt != 0 {
			et.switchToPrevTab()
			return nil
		}

		if event.Key() == tcell.KeyRight && event.Modifiers()&tcell.ModAlt != 0 {
			et.switchToNextTab()
			return nil
		}

		return event
	})
}

func (et *EditorTabs) createNewTab(name string) {
	editor := NewEditor(et.app, et.pages, et.dbApp, et.helpBar, et.statusBar)

	editor.onRunningStateChange = func(running bool) {
		et.MarkCurrentTabRunning(running)
	}

	editor.onQueryLoad = func(queryID, queryName, querySQL string) {
		et.LoadSavedQueryIntoCurrentTab(queryID, queryName, querySQL)
	}

	editor.onQuerySave = func(queryID, queryName string) {
		et.MarkCurrentTabSaved(queryID, queryName)
	}

	editor.getSQLText = func() string {
		return strings.TrimSpace(et.GetCurrentTabQuery())
	}

	editor.onCheckModified = func() {
		et.CheckCurrentTabModified()
	}

	editor.getCurrentSavedQueryID = func() string {
		return et.GetCurrentSavedQueryID()
	}

	editor.ConfigureSavedQueriesCallbacks()

	tab := &TabState{
		id:            et.nextTabID,
		name:          name,
		editor:        editor,
		modified:      false,
		running:       false,
		savedQueryID:  "",
		savedQuerySQL: "",
	}

	et.tabs = append(et.tabs, tab)
	et.nextTabID++
	et.activeTab = len(et.tabs) - 1

	et.updateDisplay()
}

func (et *EditorTabs) switchToTab(index int) {
	if index < 0 || index >= len(et.tabs) {
		return
	}
	et.activeTab = index
	et.updateDisplay()
}

func (et *EditorTabs) switchToNextTab() {
	if len(et.tabs) == 0 {
		return
	}
	et.activeTab = (et.activeTab + 1) % len(et.tabs)
	et.updateDisplay()
}

func (et *EditorTabs) switchToPrevTab() {
	if len(et.tabs) == 0 {
		return
	}
	et.activeTab = (et.activeTab - 1 + len(et.tabs)) % len(et.tabs)
	et.updateDisplay()
}

func (et *EditorTabs) updateDisplay() {
	et.renderTabBar()
	et.showActiveEditorContent()
}

func (et *EditorTabs) renderTabBar() {
	var sb strings.Builder

	for i, tab := range et.tabs {
		isActive := i == et.activeTab
		if tab.running {
			if isActive {
				sb.WriteString(fmt.Sprintf("[#%06x:#%06x:b]",
					theme.ThemeColors.Background.Hex(),
					theme.ThemeColors.Info.Hex()))
			} else {
				sb.WriteString(fmt.Sprintf("[#%06x:#%06x:-]",
					theme.ThemeColors.Foreground.Hex(),
					theme.ThemeColors.BackgroundAlt.Hex()))
			}
		} else if tab.modified {
			if isActive {
				sb.WriteString(fmt.Sprintf("[#%06x:#%06x:b]",
					theme.ThemeColors.Background.Hex(),
					theme.ThemeColors.Warning.Hex()))
			} else {
				sb.WriteString(fmt.Sprintf("[#%06x:#%06x:-]",
					theme.ThemeColors.Foreground.Hex(),
					theme.ThemeColors.BackgroundAlt.Hex()))
			}
		} else if isActive {
			sb.WriteString(fmt.Sprintf("[#%06x:#%06x:b]",
				theme.ThemeColors.Background.Hex(),
				theme.ThemeColors.Primary.Hex()))
		} else {
			sb.WriteString(fmt.Sprintf("[#%06x:#%06x:-]",
				theme.ThemeColors.Foreground.Hex(),
				theme.ThemeColors.BackgroundAlt.Hex()))
		}

		sb.WriteString(" ")

		sb.WriteString(tab.name)

		if tab.modified {
			sb.WriteString(theme.Icons.Modified)
		}

		if tab.running {
			sb.WriteString(theme.Icons.Running)
		}

		sb.WriteString("[-:-:-]")

		if i < len(et.tabs)-1 {
			sb.WriteString(" ")
		}
	}

	et.tabBar.SetText(sb.String())
}

func (et *EditorTabs) showActiveEditorContent() {
	et.contentFlex.Clear()

	if len(et.tabs) == 0 {
		return
	}

	activeEditor := et.tabs[et.activeTab].editor
	et.contentFlex.AddItem(activeEditor.mainFlex, 0, 1, true)
	et.app.SetFocus(activeEditor.sqlInput)
}

func (et *EditorTabs) GetActiveEditor() *Editor {
	if len(et.tabs) == 0 {
		return nil
	}
	return et.tabs[et.activeTab].editor
}

func (et *EditorTabs) MarkCurrentTabModified(modified bool) {
	if len(et.tabs) == 0 {
		return
	}
	et.tabs[et.activeTab].modified = modified
	et.renderTabBar()
}

func (et *EditorTabs) MarkCurrentTabRunning(running bool) {
	if len(et.tabs) == 0 {
		return
	}
	et.tabs[et.activeTab].running = running
	et.renderTabBar()
}

func (et *EditorTabs) closeCurrentTab() {
	if len(et.tabs) <= 1 {
		components.ShowError(et.pages, et.app, fmt.Errorf("cannot close the last tab"))
		return
	}

	currentTab := et.tabs[et.activeTab]

	if currentTab.modified {
		et.showCloseConfirmation(et.activeTab)
		return
	}

	et.closeTabAtIndex(et.activeTab)
}

func (et *EditorTabs) closeTabAtIndex(index int) {
	if index < 0 || index >= len(et.tabs) {
		return
	}

	et.tabs = append(et.tabs[:index], et.tabs[index+1:]...)

	if et.activeTab >= len(et.tabs) {
		et.activeTab = len(et.tabs) - 1
	}

	et.updateDisplay()
}

func (et *EditorTabs) showCloseConfirmation(tabIndex int) {
	modal := tview.NewModal().
		SetText("Tab has unsaved changes. Close anyway?").
		AddButtons([]string{"Close", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			et.pages.RemovePage("close-confirmation")
			if buttonLabel == "Close" {
				et.closeTabAtIndex(tabIndex)
			}
		})

	et.pages.AddPage("close-confirmation", modal, true, true)
}

func (et *EditorTabs) renameCurrentTab() {
	if len(et.tabs) == 0 {
		return
	}

	currentTab := et.tabs[et.activeTab]

	dialog := components.NewFormDialog(et.pages, et.app, components.FormDialogConfig{
		Title: "Rename Tab",
		Fields: []components.FormField{
			{Type: components.FieldTypeInput, Label: "Tab name", InitialValue: currentTab.name, FieldWidth: 30},
		},
		SubmitLabel: "Rename",
		CancelLabel: "Cancel",
		PageName:    "rename-tab",
		ModalWidth:  50,
		ModalHeight: 7,
		OnSubmit: func(values map[string]string) error {
			newName := values["Tab name"]
			if newName != "" {
				et.renameTab(et.activeTab, newName)
			}
			return nil
		},
	})

	dialog.Show()
}

func (et *EditorTabs) renameTab(index int, newName string) {
	if index < 0 || index >= len(et.tabs) {
		return
	}
	et.tabs[index].name = newName
	et.renderTabBar()
}

func (et *EditorTabs) Show() {
	et.pages.AddPage("editor-tabs", et.mainFlex, true, true)
	et.pages.SwitchToPage("editor-tabs")
	et.updateDisplay()
}

func (et *EditorTabs) LoadSavedQueryIntoCurrentTab(queryID, queryName, querySQL string) {
	if len(et.tabs) == 0 {
		return
	}

	currentTab := et.tabs[et.activeTab]

	currentTab.savedQueryID = queryID
	currentTab.savedQuerySQL = querySQL
	currentTab.name = queryName
	currentTab.modified = false

	currentTab.editor.sqlInput.SetText(querySQL, true)

	et.renderTabBar()
	et.app.SetFocus(currentTab.editor.sqlInput)
}

func (et *EditorTabs) GetCurrentTabQuery() string {
	if len(et.tabs) == 0 {
		return ""
	}
	return et.tabs[et.activeTab].editor.sqlInput.GetText()
}

func (et *EditorTabs) CheckCurrentTabModified() {
	if len(et.tabs) == 0 {
		return
	}

	currentTab := et.tabs[et.activeTab]
	currentSQL := strings.TrimSpace(currentTab.editor.sqlInput.GetText())

	if currentTab.savedQueryID != "" {
		savedSQL := strings.TrimSpace(currentTab.savedQuerySQL)
		modified := currentSQL != savedSQL
		if currentTab.modified != modified {
			currentTab.modified = modified
			et.renderTabBar()
		}
	} else {
		modified := currentSQL != ""
		if currentTab.modified != modified {
			currentTab.modified = modified
			et.renderTabBar()
		}
	}
}

func (et *EditorTabs) GetCurrentSavedQueryID() string {
	if len(et.tabs) == 0 {
		return ""
	}
	return et.tabs[et.activeTab].savedQueryID
}

func (et *EditorTabs) MarkCurrentTabSaved(queryID, queryName string) {
	if len(et.tabs) == 0 {
		return
	}

	currentTab := et.tabs[et.activeTab]
	currentSQL := strings.TrimSpace(currentTab.editor.sqlInput.GetText())

	currentTab.savedQueryID = queryID
	currentTab.savedQuerySQL = currentSQL
	currentTab.name = queryName
	currentTab.modified = false

	et.renderTabBar()
}
