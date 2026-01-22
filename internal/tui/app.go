package tui

import (
	"github.com/android-lewis/dbsmith/internal/app"
	"github.com/android-lewis/dbsmith/internal/tui/components"
	"github.com/android-lewis/dbsmith/internal/tui/editor"
	"github.com/android-lewis/dbsmith/internal/tui/explorer"
	"github.com/android-lewis/dbsmith/internal/tui/theme"
	"github.com/android-lewis/dbsmith/internal/tui/workspace"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TUIApp struct {
	app       *tview.Application
	pages     *tview.Pages
	dbApp     *app.App
	helpBar   *components.HelpBar
	statusBar *components.StatusBar

	workspace  *workspace.Workspace
	explorer   *explorer.Explorer
	editorTabs *editor.EditorTabs

	mainLayout *tview.Flex
}

func NewTUIApp(application *app.App) *TUIApp {
	theme.ApplyTheme()

	tuiApp := &TUIApp{
		app:       tview.NewApplication(),
		pages:     tview.NewPages(),
		dbApp:     application,
		helpBar:   components.NewHelpBar(),
		statusBar: components.NewStatusBar(application),
	}

	tuiApp.workspace = workspace.NewWorkspace(tuiApp.app, tuiApp.pages, application, tuiApp.helpBar, tuiApp.statusBar)

	tuiApp.workspace.SetConnectionSelectedCallback(func() {
		tuiApp.editorTabs = editor.NewEditorTabs(tuiApp.app, tuiApp.pages, application, tuiApp.helpBar, tuiApp.statusBar)
		tuiApp.explorer = explorer.NewExplorer(tuiApp.app, tuiApp.pages, application, tuiApp.helpBar, tuiApp.statusBar)
		tuiApp.ShowExplorer()
	})

	tuiApp.setupGlobalKeys()
	tuiApp.buildMainLayout()

	return tuiApp
}

func (t *TUIApp) setupGlobalKeys() {
	t.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF10, tcell.KeyCtrlC:
			components.ShowConfirm(t.pages, t.app, "Are you sure you want to quit?", func(confirmed bool) {
				if confirmed {
					t.app.Stop()
				}
			})
			return nil
		case tcell.KeyF2:
			t.ShowWorkspace()
			return nil
		case tcell.KeyF3:
			if t.dbApp.Connection != nil {
				t.ShowEditor()
			}
			return nil
		case tcell.KeyF4:
			if t.dbApp.Connection != nil {
				t.ShowExplorer()
			}
			return nil
		case tcell.KeyF1:
			t.helpBar.ToggleExpanded()
			t.rebuildMainLayout()
			return nil
		}
		return event
	})
}

func (t *TUIApp) buildMainLayout() {
	helpHeight := t.helpBar.GetHeight()

	t.mainLayout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(t.statusBar, 1, 0, false).
		AddItem(t.pages, 0, 1, true).
		AddItem(t.helpBar, helpHeight, 0, false)

	rootPages := tview.NewPages()
	rootPages.AddPage("main", t.mainLayout, true, true)

	t.app.SetRoot(rootPages, true)
}

func (t *TUIApp) rebuildMainLayout() {
	helpHeight := t.helpBar.GetHeight()

	t.mainLayout.Clear()
	t.mainLayout.
		AddItem(t.statusBar, 1, 0, false).
		AddItem(t.pages, 0, 1, true).
		AddItem(t.helpBar, helpHeight, 0, false)
}

func (t *TUIApp) ShowWorkspace() {
	t.helpBar.SetContext("workspace")
	t.statusBar.Update()
	t.workspace.Show()
}

func (t *TUIApp) ShowExplorer() {
	t.helpBar.SetContext("explorer")
	t.statusBar.Update()
	t.explorer.Show()
}

func (t *TUIApp) ShowEditor() {
	t.helpBar.SetContext("editor")
	t.statusBar.Update()
	t.editorTabs.Show()
}

func Run(application *app.App) error {
	tuiApp := NewTUIApp(application)

	if application.Workspace == nil || application.Connection == nil {
		tuiApp.ShowWorkspace()
	} else {
		tuiApp.ShowExplorer()
	}

	return tuiApp.app.Run()
}
