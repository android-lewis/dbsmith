package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/android-lewis/dbsmith/internal/app"
	"github.com/android-lewis/dbsmith/internal/db"
	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/tui/components"
	"github.com/android-lewis/dbsmith/internal/tui/constants"
	"github.com/android-lewis/dbsmith/internal/tui/theme"
	wsmgr "github.com/android-lewis/dbsmith/internal/workspace"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Workspace struct {
	app       *tview.Application
	pages     *tview.Pages
	dbApp     *app.App
	helpBar   *components.HelpBar
	statusBar *components.StatusBar

	mainFlex        *tview.Flex
	connectionsList *tview.List
	connections     []models.Connection
	testingConn     bool
	loadingOverlay  *components.LoadingOverlay
	connectionForm  *components.ConnectionFormManager

	workspacePath        string
	onConnectionSelected func()
}

func NewWorkspace(app *tview.Application, pages *tview.Pages, dbApp *app.App, helpBar *components.HelpBar, statusBar *components.StatusBar) *Workspace {
	w := &Workspace{
		app:       app,
		pages:     pages,
		dbApp:     dbApp,
		helpBar:   helpBar,
		statusBar: statusBar,
	}

	w.connectionForm = components.NewConnectionFormManager(pages, app)
	w.configureConnectionFormCallbacks()
	w.loadingOverlay = components.NewLoadingOverlay()

	w.buildUI()
	return w
}

func (w *Workspace) buildUI() {
	w.connectionsList = tview.NewList().
		ShowSecondaryText(true).
		SetHighlightFullLine(true).
		SetMainTextColor(theme.ThemeColors.Primary).
		SetSecondaryTextColor(theme.ThemeColors.ForegroundMuted).
		SetSelectedTextColor(theme.ThemeColors.Foreground).
		SetSelectedBackgroundColor(theme.ThemeColors.Selection)

	w.connectionsList.SetBorder(true).
		SetTitle(" Connections ").
		SetTitleAlign(tview.AlignLeft)

	theme.SetUnfocused(w.connectionsList)

	w.mainFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(w.connectionsList, 0, 1, true)

	w.testingConn = false

	w.setupKeybindings()
	w.loadConnections()
}

func (w *Workspace) setupKeybindings() {
	w.connectionsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'n', 'N':
			w.connectionForm.ShowNewConnectionForm()
			return nil
		case 'e', 'E':
			if conn := w.getSelectedConnection(); conn != nil {
				w.connectionForm.ShowEditConnectionForm(conn)
			}
			return nil
		case 'd', 'D':
			if conn := w.getSelectedConnection(); conn != nil {
				w.confirmDelete(conn.Name)
			}
			return nil
		case 't', 'T':
			if conn := w.getSelectedConnection(); conn != nil && !w.testingConn {
				w.testConnection(conn)
			}
			return nil
		}
		return event
	})
}

func (w *Workspace) getSelectedConnection() *models.Connection {
	index := w.connectionsList.GetCurrentItem()
	if index >= 0 && index < len(w.connections) {
		return &w.connections[index]
	}
	return nil
}

func (w *Workspace) loadConnections() {
	if w.dbApp.Workspace == nil {
		w.showWorkspaceSelector()
		return
	}

	w.connections = w.dbApp.Workspace.ListConnections()

	if len(w.connections) == 0 {
		w.connectionsList.Clear()
		w.connectionsList.AddItem("No connections configured", "Press N to create your first connection", 0, nil)
		return
	}

	w.renderConnectionsList()
}

func (w *Workspace) renderConnectionsList() {
	w.connectionsList.Clear()

	for i, conn := range w.connections {
		mainText, secondaryText := w.formatConnectionItem(conn)
		index := i
		w.connectionsList.AddItem(mainText, secondaryText, 0, func() {
			if index >= 0 && index < len(w.connections) {
				w.selectConnection(&w.connections[index])
			}
		})
	}
}

func (w *Workspace) formatConnectionItem(conn models.Connection) (mainText, secondaryText string) {
	statusIcon := theme.Icons.Disconnected
	if w.dbApp.Driver != nil && w.dbApp.Driver.IsConnected() {
		if w.dbApp.Connection != nil && w.dbApp.Connection.Name == conn.Name {
			statusIcon = theme.Icons.Connected
		}
	}

	mainText = fmt.Sprintf("%s %s", statusIcon, conn.Name)

	if conn.Type == models.SQLiteType {
		secondaryText = fmt.Sprintf("%s  %s", conn.Type, conn.Database)
	} else {
		secondaryText = fmt.Sprintf("%s  %s:%d/%s  user:%s", conn.Type, conn.Host, conn.Port, conn.Database, conn.Username)
	}

	return mainText, secondaryText
}

func (w *Workspace) testConnection(conn *models.Connection) {
	w.testingConn = true

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutSchemaLoad)
		defer cancel()

		driverFactory := db.NewDriverFactory()
		driver, err := driverFactory.Create(conn)
		if err != nil {
			w.app.QueueUpdateDraw(func() {
				w.testingConn = false
				components.ShowError(
					w.pages,
					w.app,
					fmt.Errorf("connection test failed: %v", err),
				)
			})
			return
		}

		err = driver.Connect(ctx, conn, w.dbApp.SecretsManager)
		if err != nil {
			w.app.QueueUpdateDraw(func() {
				w.testingConn = false
				components.ShowError(
					w.pages,
					w.app,
					fmt.Errorf("connection test failed: %v", err),
				)
			})
			return
		}

		_ = driver.Disconnect(ctx)

		w.app.QueueUpdateDraw(func() {
			w.testingConn = false
			components.ShowInfo(
				w.pages,
				w.app,
				fmt.Sprintf("Connection '%s' tested successfully!", conn.Name),
			)
		})
	}()
}

func (w *Workspace) showWorkspaceSelector() {
	var form *tview.Form

	form = tview.NewForm().
		AddInputField("Workspace File", "", 50, nil, nil).
		AddButton("Create New", func() {
			input := form.GetFormItem(0).(*tview.InputField)
			path := input.GetText()
			if path == "" {
				components.ShowError(w.pages, w.app, fmt.Errorf("workspace file path is required"))
				return
			}
			w.createWorkspace(path)
		}).
		AddButton("Load Existing", func() {
			input := form.GetFormItem(0).(*tview.InputField)
			path := input.GetText()
			if path == "" {
				components.ShowError(w.pages, w.app, fmt.Errorf("workspace file path is required"))
				return
			}
			w.loadWorkspace(path)
		}).
		AddButton("Cancel", func() {
			w.app.Stop()
		})

	form.SetBorder(true).
		SetTitle(" Workspace Manager ").
		SetTitleAlign(tview.AlignCenter)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			w.app.Stop()
		}
		return event
	})

	w.pages.AddPage("workspace-selector", form, true, true)
}

func (w *Workspace) createWorkspace(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		components.ShowError(w.pages, w.app, err)
		return
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		components.ShowError(w.pages, w.app, fmt.Errorf("failed to create directory: %w", err))
		return
	}

	ws := wsmgr.New()
	ws.SetName("Default Workspace")

	if err := ws.Save(absPath); err != nil {
		components.ShowError(w.pages, w.app, err)
		return
	}

	w.dbApp.Workspace = ws
	w.workspacePath = absPath
	w.pages.RemovePage("workspace-selector")
	w.loadConnections()
	components.ShowInfo(w.pages, w.app, fmt.Sprintf("Workspace created: %s", absPath))
}

func (w *Workspace) loadWorkspace(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		components.ShowError(w.pages, w.app, err)
		return
	}

	ws, err := wsmgr.Load(absPath)
	if err != nil {
		components.ShowError(w.pages, w.app, err)
		return
	}

	w.dbApp.Workspace = ws
	w.workspacePath = absPath
	w.pages.RemovePage("workspace-selector")
	w.loadConnections()
}

func (w *Workspace) configureConnectionFormCallbacks() {
	w.connectionForm.SetCallbacks(
		func(conn models.Connection, password string, isEdit bool) error {
			if password != "" {
				if err := w.dbApp.SecretsManager.StoreSecret(conn.SecretKeyID, password); err != nil {
					return fmt.Errorf("failed to store password: %w", err)
				}
			}

			var err error
			if isEdit {
				err = w.dbApp.Workspace.UpdateConnection(conn)
			} else {
				err = w.dbApp.Workspace.AddConnection(conn)
			}

			if err != nil {
				return err
			}

			if w.workspacePath == "" {
				return fmt.Errorf("workspace file path not set")
			}

			if err := w.dbApp.Workspace.Save(w.workspacePath); err != nil {
				return err
			}

			w.loadConnections()
			return nil
		},
		func(message string) {
			components.ShowInfo(w.pages, w.app, message)
		},
	)
}

func (w *Workspace) confirmDelete(name string) {
	components.ShowConfirm(w.pages, w.app, fmt.Sprintf("Delete connection '%s'?", name), func(confirmed bool) {
		if !confirmed {
			return
		}

		if err := w.dbApp.Workspace.DeleteConnection(name); err != nil {
			components.ShowError(w.pages, w.app, err)
			return
		}

		if err := w.dbApp.Workspace.Save(w.workspacePath); err != nil {
			components.ShowError(w.pages, w.app, err)
			return
		}

		w.loadConnections()
		components.ShowInfo(w.pages, w.app, fmt.Sprintf("Connection deleted: %s", name))
	})
}

func (w *Workspace) SetConnectionSelectedCallback(callback func()) {
	w.onConnectionSelected = callback
}

func (w *Workspace) selectConnection(conn *models.Connection) {
	w.loadingOverlay.Show("Connecting...", true)
	w.pages.AddPage("loading", w.loadingOverlay, true, true)

	go func() {
		if err := w.dbApp.ConnectToDatabase(conn); err != nil {
			w.app.QueueUpdateDraw(func() {
				w.pages.RemovePage("loading")
				components.ShowError(w.pages, w.app, fmt.Errorf("failed to connect: %w", err))
			})
			return
		}

		w.app.QueueUpdateDraw(func() {
			w.pages.RemovePage("loading")
			w.statusBar.Update()
			w.renderConnectionsList()

			if w.onConnectionSelected != nil {
				w.onConnectionSelected()
			}
		})
	}()
}

func (w *Workspace) Show() {
	w.pages.RemovePage("workspace")
	w.pages.AddPage("workspace", w.mainFlex, true, true)
	w.pages.SwitchToPage("workspace")
	w.app.SetFocus(w.connectionsList)
}
