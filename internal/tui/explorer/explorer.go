package explorer

import (
	"context"
	"fmt"

	"github.com/android-lewis/dbsmith/internal/app"
	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/tui/components"
	"github.com/android-lewis/dbsmith/internal/tui/constants"
	"github.com/android-lewis/dbsmith/internal/tui/theme"
	"github.com/android-lewis/dbsmith/internal/tui/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Explorer struct {
	app       *tview.Application
	pages     *tview.Pages
	dbApp     *app.App
	helpBar   *components.HelpBar
	statusBar *components.StatusBar

	mainFlex     *tview.Flex
	leftFlex     *tview.Flex
	schemasList  *tview.List
	tablesList   *tview.List
	schemaFlex   *tview.Flex
	columnsTable *tview.Table
	indexTable   *tview.Table
	dataTable    *tview.Table

	selectedSchema  string
	selectedTable   string
	showSchemas     bool
	showIndexes     bool
	showDataPreview bool
	dataLimit       int
	dataOffset      int
	focusedPanel    int
}

const (
	panelSchemas = iota
	panelTables
	panelSchema
	panelIndexes
	panelData
)

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

func (e *Explorer) Show() {
	e.pages.AddPage("explorer", e.mainFlex, true, true)
	e.pages.SwitchToPage("explorer")
	e.focusedPanel = panelSchemas
	e.updateFocus()
	e.loadSchemas()
}

func (e *Explorer) buildUI() {
	e.schemasList = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)

	e.schemasList.SetBorder(true).
		SetTitle(" Schemas ").
		SetTitleAlign(tview.AlignLeft)

	e.schemasList.SetSelectedBackgroundColor(theme.ThemeColors.Selection).
		SetSelectedTextColor(theme.ThemeColors.SelectionText)

	e.tablesList = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)

	e.tablesList.SetBorder(true).
		SetTitle(" Tables ").
		SetTitleAlign(tview.AlignLeft)

	e.tablesList.SetSelectedBackgroundColor(theme.ThemeColors.Selection).
		SetSelectedTextColor(theme.ThemeColors.SelectionText)

	e.leftFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(e.schemasList, 0, 1, true).
		AddItem(e.tablesList, 0, 2, false)

	e.columnsTable = tview.NewTable().
		SetBorders(false).
		SetSelectable(false, false)

	e.columnsTable.SetBorder(true).
		SetTitle(" Columns ").
		SetTitleAlign(tview.AlignLeft)

	e.indexTable = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)

	e.indexTable.SetBorder(true).
		SetTitle(" Indexes ").
		SetTitleAlign(tview.AlignLeft)

	e.indexTable.SetSelectedStyle(tcell.StyleDefault.
		Background(theme.ThemeColors.Selection).
		Foreground(theme.ThemeColors.SelectionText))

	e.schemaFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(e.columnsTable, 0, 1, false)

	e.dataTable = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)

	e.dataTable.SetBorder(true).
		SetTitle(" Data Preview ").
		SetTitleAlign(tview.AlignLeft)

	e.dataTable.SetSelectedStyle(tcell.StyleDefault.
		Background(theme.ThemeColors.Selection).
		Foreground(theme.ThemeColors.SelectionText))

	e.mainFlex = tview.NewFlex().
		AddItem(e.leftFlex, 25, 0, true).
		AddItem(e.schemaFlex, 35, 0, false).
		AddItem(e.dataTable, 0, 1, false)

	e.setupKeybindings()
}

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

func (e *Explorer) updateFocus() {
	theme.SetUnfocused(e.schemasList)
	theme.SetUnfocused(e.tablesList)
	theme.SetUnfocused(e.columnsTable)
	theme.SetUnfocused(e.indexTable)
	theme.SetUnfocused(e.dataTable)

	switch e.focusedPanel {
	case panelSchemas:
		e.setFocusedPrimative(e.schemasList)
	case panelTables:
		e.setFocusedPrimative(e.tablesList)
	case panelSchema:
		e.setFocusedPrimative(e.columnsTable)
	case panelIndexes:
		e.setFocusedPrimative(e.indexTable)
	case panelData:
		e.setFocusedPrimative(e.dataTable)
	}
}

func (e *Explorer) setFocusedPrimative(p tview.Primitive) {
	theme.SetFocused(p)
	e.app.SetFocus(p)
}

func (e *Explorer) loadSchemas() {
	e.schemasList.Clear()

	if e.dbApp.Explorer == nil {
		e.schemasList.AddItem("No connection", "", 0, nil)
		return
	}

	e.schemasList.AddItem("Loading...", "", 0, nil)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutSchemaLoad)
		defer cancel()

		schemas, err := e.dbApp.Explorer.GetSchemas(ctx)
		if err != nil {
			e.app.QueueUpdateDraw(func() {
				e.schemasList.Clear()
				e.schemasList.AddItem("Error loading schemas", "", 0, nil)
				components.ShowError(e.pages, e.app, fmt.Errorf("failed to load schemas: %w", err))
			})
			return
		}

		e.app.QueueUpdateDraw(func() {
			e.schemasList.Clear()

			if len(schemas) == 0 {
				e.schemasList.AddItem("No schemas", "", 0, nil)
				return
			}

			for _, schema := range schemas {
				localSchemaName := schema.Name
				e.schemasList.AddItem(localSchemaName, "", 0, func() {
					e.loadTablesForSchema(localSchemaName)
				})
			}

			if len(schemas) > 0 {
				e.schemasList.SetCurrentItem(0)
				e.selectedSchema = schemas[0].Name
				e.loadTablesForSchema(schemas[0].Name)
			}
		})
	}()
}

func (e *Explorer) loadTablesForSchema(schemaName string) {
	e.selectedSchema = schemaName
	e.tablesList.Clear()

	if e.dbApp.Explorer == nil {
		e.tablesList.AddItem("No connection", "", 0, nil)
		return
	}

	e.tablesList.AddItem("Loading...", "", 0, nil)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutSchemaLoad)
		defer cancel()

		schema := models.Schema{Name: schemaName}
		tables, err := e.dbApp.Explorer.GetTables(ctx, schema)
		if err != nil {
			e.app.QueueUpdateDraw(func() {
				e.tablesList.Clear()
				e.tablesList.AddItem("Error loading tables", "", 0, nil)
				components.ShowError(e.pages, e.app, fmt.Errorf("failed to load tables: %w", err))
			})
			return
		}

		e.app.QueueUpdateDraw(func() {
			e.tablesList.Clear()

			if len(tables) == 0 {
				e.tablesList.AddItem("No tables", "", 0, nil)
				return
			}

			for _, table := range tables {
				localTableName := table.Name
				e.tablesList.AddItem(localTableName, "", 0, func() {
					e.loadTableDetails(localTableName)
				})
			}

			if len(tables) > 0 {
				e.tablesList.SetCurrentItem(0)
				e.loadTableDetails(tables[0].Name)
			}
		})
	}()
}

func (e *Explorer) loadTableDetails(tableName string) {
	e.selectedTable = tableName
	e.dataOffset = 0

	go e.loadColumns(tableName)

	if e.showDataPreview {
		go e.loadDataPreview(tableName)
	}

	if e.showIndexes {
		go e.loadIndexes(tableName)
	}
}

func (e *Explorer) loadColumns(tableName string) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutSchemaLoad)
	defer cancel()

	schema, err := e.dbApp.Explorer.GetTableColumns(ctx, tableName)
	if err != nil {
		e.app.QueueUpdateDraw(func() {
			e.columnsTable.Clear()
			e.columnsTable.SetContent(nil)
			e.columnsTable.SetCell(0, 0, tview.NewTableCell("Error loading columns"))
		})
		return
	}

	e.app.QueueUpdateDraw(func() {
		e.columnsTable.Clear()

		content := components.NewColumnContent(schema.Columns)
		content.ApplyAlternatingRowColors()
		e.columnsTable.SetContent(content)
	})
}

func (e *Explorer) loadIndexes(tableName string) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutSchemaLoad)
	defer cancel()

	indexes, err := e.dbApp.Explorer.GetTableIndexes(ctx, tableName)
	if err != nil {
		e.app.QueueUpdateDraw(func() {
			e.indexTable.Clear()
			e.indexTable.SetContent(nil)
			e.indexTable.SetCell(0, 0, tview.NewTableCell("Error loading indexes"))
		})
		return
	}

	e.app.QueueUpdateDraw(func() {
		e.indexTable.Clear()

		if len(indexes) == 0 {
			e.indexTable.SetContent(nil)
			e.indexTable.SetCell(0, 0, tview.NewTableCell("No indexes"))
			return
		}

		content := components.NewIndexContent(indexes)
		content.ApplyAlternatingRowColors()
		e.indexTable.SetContent(content)
	})
}

func (e *Explorer) loadDataPreview(tableName string) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutSchemaLoad)
	defer cancel()

	result, err := e.dbApp.Explorer.GetTableData(ctx, tableName, e.dataLimit, e.dataOffset)
	if err != nil {
		e.app.QueueUpdateDraw(func() {
			e.dataTable.Clear()
			e.dataTable.SetContent(nil)
			e.dataTable.SetCell(0, 0, tview.NewTableCell(fmt.Sprintf("Error: %v", err)))
		})
		return
	}

	e.app.QueueUpdateDraw(func() {
		e.dataTable.Clear()

		startRow := e.dataOffset + 1
		endRow := e.dataOffset + len(result.Rows)
		title := fmt.Sprintf(" Data Preview [%s-%s of %s] ",
			utils.FormatNumber(int64(startRow)),
			utils.FormatNumber(int64(endRow)),
			utils.FormatNumber(result.RowCount))
		e.dataTable.SetTitle(title)

		content := components.NewQueryResultContentWithMaxLen(result, constants.MaxPreviewCellLen)
		content.ApplyAlternatingRowColors()
		e.dataTable.SetContent(content)
	})
}

func (e *Explorer) toggleSchemas() {
	e.showSchemas = !e.showSchemas

	if !e.showSchemas && e.focusedPanel == panelSchemas {
		e.focusedPanel = panelTables
	}

	e.leftFlex.Clear()
	if e.showSchemas {
		e.leftFlex.
			AddItem(e.schemasList, 0, 1, false).
			AddItem(e.tablesList, 0, 2, false)

		if len(e.schemasList.FindItems("", "", false, false)) == 0 {
			go e.loadSchemas()
		}
	} else {
		e.leftFlex.AddItem(e.tablesList, 0, 1, false)
	}

	e.updateFocus()
}

func (e *Explorer) toggleIndexes() {
	e.showIndexes = !e.showIndexes

	if !e.showIndexes && e.focusedPanel == panelIndexes {
		e.focusedPanel = panelSchema
	}

	e.schemaFlex.Clear()
	if e.showIndexes {
		e.schemaFlex.
			AddItem(e.columnsTable, 0, 2, false).
			AddItem(e.indexTable, 0, 1, false)

		if e.selectedTable != "" {
			go e.loadIndexes(e.selectedTable)
		}
	} else {
		e.schemaFlex.AddItem(e.columnsTable, 0, 1, false)
	}

	e.updateFocus()
}

func (e *Explorer) toggleDataPreview() {
	e.showDataPreview = !e.showDataPreview

	if !e.showDataPreview && e.focusedPanel == panelData {
		e.focusedPanel = panelTables
	}

	if e.showDataPreview && e.selectedTable != "" {
		go e.loadDataPreview(e.selectedTable)
	}

	e.rebuildLayout()
}

func (e *Explorer) rebuildLayout() {
	e.mainFlex.Clear()

	if e.showDataPreview {
		e.mainFlex.
			AddItem(e.leftFlex, 25, 0, false).
			AddItem(e.schemaFlex, 35, 0, false).
			AddItem(e.dataTable, 0, 1, false)
	} else {
		e.mainFlex.
			AddItem(e.leftFlex, 25, 0, false).
			AddItem(e.schemaFlex, 0, 1, false)
	}

	e.updateFocus()
}

func (e *Explorer) showServerInfo() {
	ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutSchemaLoad)
	defer cancel()

	info, err := e.dbApp.Driver.GetServerInfo(ctx)
	if err != nil {
		e.statusBar.Update()
		return
	}

	// Build the content
	content := fmt.Sprintf(`[::b]%s Server Information[::-]

[yellow]Version:[white]     %s
[yellow]Database:[white]    %s
[yellow]User:[white]        %s
[yellow]Uptime:[white]      %s
[yellow]Size:[white]        %s

[yellow]Connections:[white] %d / %d`,
		info.ServerType,
		info.Version,
		info.CurrentDatabase,
		info.CurrentUser,
		info.Uptime,
		info.DatabaseSize,
		info.ConnectionCount,
		info.MaxConnections,
	)

	// Add additional info
	if len(info.AdditionalInfo) > 0 {
		content += "\n\n[::b]Additional Info[::-]"
		for key, value := range info.AdditionalInfo {
			content += fmt.Sprintf("\n[yellow]%s:[white] %s", key, value)
		}
	}

	// Create modal
	modal := tview.NewTextView().
		SetDynamicColors(true).
		SetText(content)
	modal.SetBorder(true).
		SetTitle(" Server Info ").
		SetTitleAlign(tview.AlignCenter)

	// Create frame with padding
	frame := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(modal, 18, 0, true).
			AddItem(nil, 0, 1, false), 60, 0, true).
		AddItem(nil, 0, 1, false)

	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Rune() == 's' || event.Rune() == 'S' {
			e.pages.RemovePage("serverinfo")
			e.updateFocus()
			return nil
		}
		return event
	})

	e.pages.AddPage("serverinfo", frame, true, true)
	e.app.SetFocus(frame)
}
