package editor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/android-lewis/dbsmith/internal/app"
	querysafety "github.com/android-lewis/dbsmith/internal/editor"
	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/tui/components"
	"github.com/android-lewis/dbsmith/internal/tui/constants"
	"github.com/android-lewis/dbsmith/internal/tui/theme"
	"github.com/android-lewis/dbsmith/internal/tui/utils"
	"github.com/android-lewis/dbsmith/internal/util"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type editorMode int

const (
	modeExecute editorMode = iota
	modeAnalyze
)

type Editor struct {
	app       *tview.Application
	pages     *tview.Pages
	dbApp     *app.App
	helpBar   *components.HelpBar
	statusBar *components.StatusBar

	mainFlex     *tview.Flex
	sqlInput     *components.SQLEditor
	resultsTable *tview.Table
	analysisView *tview.TextView
	bottomFlex   *tview.Flex
	queryStats   *components.QueryStats

	mode           editorMode
	lastResult     *models.QueryResult
	queryCancel    context.CancelFunc
	isQueryRunning bool

	// Result state for selection callback (avoids re-registering callback per query)
	resultRowCount int
	resultExecMs   int64

	savedQueriesManager    *components.SavedQueriesManager
	exportManager          *components.ExportManager
	onRunningStateChange   func(bool)
	onQueryLoad            func(queryID, queryName, querySQL string)
	onQuerySave            func(queryID, queryName string)
	onCheckModified        func()
	getSQLText             func() string
	getCurrentSavedQueryID func() string
}

func NewEditor(app *tview.Application, pages *tview.Pages, dbApp *app.App, helpBar *components.HelpBar, statusBar *components.StatusBar) *Editor {
	e := &Editor{
		app:       app,
		pages:     pages,
		dbApp:     dbApp,
		helpBar:   helpBar,
		statusBar: statusBar,
		mode:      modeExecute,
	}

	e.savedQueriesManager = components.NewSavedQueriesManager(pages, app, dbApp.Workspace)
	e.exportManager = components.NewExportManager(pages, app)

	e.buildUI()
	e.configureExportCallbacks()
	return e
}

func (e *Editor) buildUI() {
	e.sqlInput = components.NewSQLEditor().
		SetPlaceholder("Enter SQL query here...\nPress F5 or Shift+Enter to execute, Tab to autocomplete")

	if e.dbApp.Executor != nil && e.dbApp.Executor.GetDriver() != nil {
		driver := e.dbApp.Executor.GetDriver()
		if conn := driver.GetConnection(); conn != nil {
			e.sqlInput.SetDialect(conn.GetSQLDialect())
		}
	}

	dialect := e.getDialectDisplayName()
	e.sqlInput.SetBorder(true).
		SetTitle(fmt.Sprintf(" SQL Editor [%s] ", dialect)).
		SetTitleAlign(tview.AlignLeft)

	e.sqlInput.SetChangedFunc(func() {
		if e.onCheckModified != nil {
			e.onCheckModified()
		}
	})

	e.resultsTable = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)

	e.resultsTable.SetBorder(true).
		SetTitle(" Results ").
		SetTitleAlign(tview.AlignLeft)

	e.resultsTable.SetSelectedStyle(tcell.StyleDefault.
		Background(theme.ThemeColors.Selection).
		Foreground(theme.ThemeColors.SelectionText))

	// Register selection callback once (uses state fields updated by displayResults)
	e.resultsTable.SetSelectionChangedFunc(func(row, col int) {
		if e.resultRowCount > 0 && row > 0 && row <= e.resultRowCount {
			e.resultsTable.SetTitle(fmt.Sprintf(" Results [Row %s of %s, %dms] ",
				utils.FormatNumber(int64(row)),
				utils.FormatNumber(int64(e.resultRowCount)),
				e.resultExecMs))
		} else if e.resultRowCount > 0 {
			e.resultsTable.SetTitle(fmt.Sprintf(" Results [%s rows, %dms] ",
				utils.FormatNumber(int64(e.resultRowCount)),
				e.resultExecMs))
		}
	})

	e.analysisView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)

	e.analysisView.SetBorder(true).
		SetTitle(" Query Analysis ").
		SetTitleAlign(tview.AlignLeft)

	e.queryStats = components.NewQueryStats()

	e.bottomFlex = tview.NewFlex().
		AddItem(e.resultsTable, 0, 1, false)

	e.mainFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(e.sqlInput, 8, 2, true).
		AddItem(e.queryStats, 3, 0, false).
		AddItem(e.bottomFlex, 0, 3, false)

	e.setupKeybindings()
}

func (e *Editor) setupKeybindings() {
	e.sqlInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			if e.isQueryRunning && e.queryCancel != nil {
				e.queryCancel()
			}
			return nil

		case tcell.KeyF5:
			e.executeQuery()
			return nil

		case tcell.KeyEnter:
			if event.Modifiers()&tcell.ModShift != 0 {
				e.executeQuery()
				return nil
			}
		}

		if event.Modifiers()&tcell.ModAlt != 0 {
			switch event.Rune() {
			case 'm', 'M':
				e.toggleMode()
				return nil
			case 's', 'S':
				e.savedQueriesManager.QuickSave(e.sqlInput)
				return nil
			case 'l', 'L':
				e.savedQueriesManager.ShowLoadDialog(e.sqlInput, func(sql string) {
					e.sqlInput.SetText(sql, true)
				})
				return nil
			case 'e', 'E':
				if e.lastResult != nil {
					e.exportManager.ShowExportDialog()
				}
				return nil
			}
		}

		if event.Key() == tcell.KeyRune && event.Rune() == 'S' && event.Modifiers()&tcell.ModAlt != 0 && event.Modifiers()&tcell.ModShift != 0 {
			e.savedQueriesManager.ShowSaveDialog(e.sqlInput)
			return nil
		}

		return event
	})

	e.resultsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			e.app.SetFocus(e.sqlInput)
			theme.SetFocused(e.sqlInput)
			theme.SetUnfocused(e.resultsTable)
			return nil
		}

		if event.Modifiers()&tcell.ModAlt != 0 && (event.Rune() == 'e' || event.Rune() == 'E') {
			if e.lastResult != nil {
				e.exportManager.ShowExportDialog()
			}
			return nil
		}

		return event
	})
}

func (e *Editor) getDialectDisplayName() string {
	if e.dbApp.Executor != nil && e.dbApp.Executor.GetDriver() != nil {
		driver := e.dbApp.Executor.GetDriver()
		if conn := driver.GetConnection(); conn != nil {
			return util.DialectDisplayName(string(conn.Type))
		}
	}
	return "SQL"
}

func (e *Editor) ConfigureSavedQueriesCallbacks() {
	e.savedQueriesManager.SetCallbacks(
		e.onQuerySave,
		e.onQueryLoad,
		e.getSQLText,
		e.getCurrentSavedQueryID,
	)
}

func (e *Editor) configureExportCallbacks() {
	e.exportManager.SetCallbacks(
		func() *models.QueryResult {
			return e.lastResult
		},
		func() string {
			if e.dbApp.Executor != nil && e.dbApp.Executor.GetDriver() != nil {
				driver := e.dbApp.Executor.GetDriver()
				if conn := driver.GetConnection(); conn != nil {
					return string(conn.Type)
				}
			}
			return ""
		},
	)
}

func (e *Editor) executeQuery() {
	sql, err := e.validateQueryPrerequisites()
	if err != nil {
		components.ShowError(e.pages, e.app, err)
		return
	}

	// Check for destructive queries and show confirmation
	safetyInfo := querysafety.AnalyzeQuerySafety(sql)
	if safetyInfo.IsDestructive {
		confirmMsg := fmt.Sprintf("%s Query Warning\n\n%s\n\nAre you sure you want to execute this query?",
			safetyInfo.QueryType, safetyInfo.Warning)

		components.ShowConfirm(e.pages, e.app, confirmMsg, func(confirmed bool) {
			if confirmed {
				e.prepareForQueryExecution()
				go e.runQuery(sql)
			}
		})
		return
	}

	e.prepareForQueryExecution()
	go e.runQuery(sql)
}

func (e *Editor) validateQueryPrerequisites() (string, error) {
	sql := strings.TrimSpace(e.sqlInput.GetText())
	if sql == "" {
		return "", fmt.Errorf("no SQL query to execute")
	}

	if e.dbApp.Executor == nil {
		return "", fmt.Errorf("no database connection")
	}

	return sql, nil
}

func (e *Editor) prepareForQueryExecution() {
	e.resultsTable.Clear()
	e.resultsTable.SetCell(0, 0, tview.NewTableCell("Executing query..."))

	dialectName := e.getDialectDisplayName()
	e.sqlInput.SetTitle(fmt.Sprintf(" SQL Editor [%s] Running - Press Esc to cancel ", dialectName))

	e.isQueryRunning = true
	if e.onRunningStateChange != nil {
		e.onRunningStateChange(true)
	}
}

func (e *Editor) runQuery(sql string) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutQueryExec)
	e.queryCancel = cancel
	cancelled := false

	defer e.cleanupAfterQuery(&cancelled)

	if e.mode == modeExecute {
		cancelled = e.executeQueryMode(ctx, sql)
	} else {
		cancelled = e.executeAnalyzeMode(ctx, sql)
	}
}

func (e *Editor) cleanupAfterQuery(cancelled *bool) {
	e.isQueryRunning = false
	if e.onRunningStateChange != nil {
		e.app.QueueUpdateDraw(func() {
			e.onRunningStateChange(false)
		})
	}
	e.queryCancel = nil

	dialectName := e.getDialectDisplayName()
	e.app.QueueUpdateDraw(func() {
		e.sqlInput.SetTitle(fmt.Sprintf(" SQL Editor [%s] ", dialectName))
	})

	if *cancelled {
		e.showCancellationMessage()
	}
}

func (e *Editor) showCancellationMessage() {
	switch e.mode {
	case modeExecute:
		e.app.QueueUpdateDraw(func() {
			e.resultsTable.Clear()
			e.resultsTable.SetCell(0, 0,
				tview.NewTableCell("Query cancelled by user").
					SetTextColor(theme.ThemeColors.Warning))
		})
	case modeAnalyze:
		e.app.QueueUpdateDraw(func() {
			e.analysisView.Clear()
			e.analysisView.SetText("Analysis cancelled by user")
		})
	}
}

func (e *Editor) executeQueryMode(ctx context.Context, sql string) bool {
	startTime := time.Now()
	result, err := e.dbApp.Executor.ExecuteQuery(ctx, sql)
	duration := time.Since(startTime)

	if err != nil {
		if ctx.Err() == context.Canceled {
			return true
		}
		e.app.QueueUpdateDraw(func() {
			components.ShowError(e.pages, e.app, fmt.Errorf("query failed: %w", err))
			e.resultsTable.Clear()
		})
		return false
	}

	e.lastResult = result
	rowCount := len(result.Rows)

	e.app.QueueUpdateDraw(func() {
		e.queryStats.RecordQuery(duration, rowCount)
		e.displayResults(result)
	})
	return false
}

func (e *Editor) executeAnalyzeMode(ctx context.Context, sql string) bool {
	result, err := e.dbApp.Executor.GetQueryExecutionPlan(ctx, sql)
	if err != nil {
		if ctx.Err() == context.Canceled {
			return true
		}
		e.app.QueueUpdateDraw(func() {
			components.ShowError(e.pages, e.app, fmt.Errorf("analysis failed: %w", err))
			e.analysisView.Clear()
		})
		return false
	}

	e.app.QueueUpdateDraw(func() {
		e.displayAnalysis(result)
	})
	return false
}

func (e *Editor) displayResults(result *models.QueryResult) {
	e.resultsTable.Clear()

	rowCount := len(result.Rows)
	rowCountStr := utils.FormatNumber(int64(rowCount))

	// Update state for selection callback (registered once in buildUI)
	e.resultRowCount = rowCount
	e.resultExecMs = result.ExecutionMs

	if rowCount == 0 {
		e.resultsTable.SetTitle(fmt.Sprintf(" Results [0 rows, %dms] ", result.ExecutionMs))
		e.resultsTable.SetContent(nil)
		e.resultsTable.SetCell(0, 0, tview.NewTableCell("Query executed successfully, no rows returned"))
		return
	}

	e.resultsTable.SetTitle(fmt.Sprintf(" Results [%s rows, %dms] ", rowCountStr, result.ExecutionMs))

	content := components.NewQueryResultContent(result)
	content.ApplyAlternatingRowColors()
	e.resultsTable.SetContent(content)

	e.app.SetFocus(e.resultsTable)
	theme.SetUnfocused(e.sqlInput)
	theme.SetFocused(e.resultsTable)
}

func (e *Editor) displayAnalysis(result *models.QueryResult) {
	e.analysisView.Clear()

	var output strings.Builder
	output.WriteString(theme.ColorTag(theme.ColorWarning, true))
	output.WriteString("Query Plan")
	output.WriteString(theme.ColorTagReset())
	output.WriteString("\n\n")

	for _, row := range result.Rows {
		for _, val := range row {
			output.WriteString(fmt.Sprintf("%v\n", val))
		}
	}

	e.analysisView.SetText(output.String())
}

func (e *Editor) toggleMode() {
	if e.mode == modeExecute {
		e.mode = modeAnalyze
		e.bottomFlex.Clear()
		e.bottomFlex.AddItem(e.analysisView, 0, 1, false)
		e.analysisView.Clear()
		e.analysisView.SetText(theme.ColorTag(theme.ColorForegroundMuted, false) + "Execute a query to see the query plan" + theme.ColorTagReset())
		components.ShowInfo(e.pages, e.app, "Switched to Analyze mode")
	} else {
		e.mode = modeExecute
		e.bottomFlex.Clear()
		e.bottomFlex.AddItem(e.resultsTable, 0, 1, false)
		components.ShowInfo(e.pages, e.app, "Switched to Execute mode")
	}
}

func (e *Editor) Show() {
	e.pages.AddPage("editor", e.mainFlex, true, true)
	e.pages.SwitchToPage("editor")
	e.app.SetFocus(e.sqlInput)
	theme.SetFocused(e.sqlInput)
}
