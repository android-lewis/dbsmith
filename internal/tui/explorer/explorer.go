package explorer

import (
	"github.com/android-lewis/dbsmith/internal/app"
	"github.com/android-lewis/dbsmith/internal/tui/components"
	"github.com/android-lewis/dbsmith/internal/tui/constants"
	"github.com/rivo/tview"
)

// Explorer provides a TUI for browsing database schemas, tables, and data.
type Explorer struct {
	app       *tview.Application
	pages     *tview.Pages
	dbApp     *app.App
	helpBar   *components.HelpBar
	statusBar *components.StatusBar

	// Layout containers
	mainFlex   *tview.Flex
	leftFlex   *tview.Flex
	schemaFlex *tview.Flex

	// UI widgets
	schemasList  *tview.List
	tablesList   *tview.List
	columnsTable *tview.Table
	indexTable   *tview.Table
	dataTable    *tview.Table

	// State
	selectedSchema  string
	selectedTable   string
	showSchemas     bool
	showIndexes     bool
	showDataPreview bool
	dataLimit       int
	dataOffset      int
	focusedPanel    int
}

// NewExplorer creates a new Explorer instance
func NewExplorer(app *tview.Application, pages *tview.Pages, dbApp *app.App, helpBar *components.HelpBar, statusBar *components.StatusBar) *Explorer {
	e := &Explorer{
		app:             app,
		pages:           pages,
		dbApp:           dbApp,
		helpBar:         helpBar,
		statusBar:       statusBar,
		dataLimit:       constants.DefaultDataLimit,
		showDataPreview: true,
		showSchemas:     true,
	}

	e.buildUI()
	return e
}

// Show displays the explorer view
func (e *Explorer) Show() {
	e.pages.AddPage("explorer", e.mainFlex, true, true)
	e.pages.SwitchToPage("explorer")
	e.focusedPanel = panelSchemas
	e.updateFocus()
	e.loadSchemas()
}

// buildUI constructs all UI components and layout
func (e *Explorer) buildUI() {
	// Build individual widgets
	e.schemasList = e.buildSchemasList()
	e.tablesList = e.buildTablesList()
	e.columnsTable = e.buildColumnsTable()
	e.indexTable = e.buildIndexTable()
	e.dataTable = e.buildDataTable()

	// Build left panel (schemas + tables)
	e.leftFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(e.schemasList, 0, 1, true).
		AddItem(e.tablesList, 0, 2, false)

	// Build schema detail panel (columns + indexes)
	e.schemaFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(e.columnsTable, 0, 1, false)

	// Build main layout
	e.mainFlex = tview.NewFlex().
		AddItem(e.leftFlex, 25, 0, true).
		AddItem(e.schemaFlex, 35, 0, false).
		AddItem(e.dataTable, 0, 1, false)

	e.setupKeybindings()
}
