package editor

import (
	"context"
	"time"

	"github.com/android-lewis/dbsmith/internal/autocomplete"
	"github.com/android-lewis/dbsmith/internal/logging"
	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/tui/components"
	"github.com/android-lewis/dbsmith/internal/tui/constants"
	"github.com/gdamore/tcell/v2"
)

type completionState struct {
	items    []autocomplete.Item
	selected int
	replace  autocomplete.Range
	active   bool
}

type completionCache struct {
	context  autocomplete.Context
	loadedAt time.Time
}

const completionCacheTTL = 30 * time.Second
const completionMaxTables = 200

func (e *Editor) handleCompletionInput(event *tcell.EventKey) bool {
	if e.completionState.active {
		switch event.Key() {
		case tcell.KeyUp:
			e.selectPrevCompletion()
			return true
		case tcell.KeyDown:
			e.selectNextCompletion()
			return true
		case tcell.KeyTab:
			e.selectNextCompletion()
			return true
		case tcell.KeyBacktab:
			e.selectPrevCompletion()
			return true
		case tcell.KeyEnter:
			if event.Modifiers()&tcell.ModShift != 0 {
				e.hideCompletion()
				e.executeQuery()
				return true
			}
			e.applyCompletion()
			return true
		case tcell.KeyEscape:
			e.hideCompletion()
			return true
		case tcell.KeyF5:
			e.hideCompletion()
			e.executeQuery()
			return true
		default:
			e.hideCompletion()
			return false
		}
	}

	if event.Key() == tcell.KeyTab {
		if e.dbApp != nil && e.dbApp.Workspace != nil && !e.dbApp.Workspace.GetAutocompleteEnabled() {
			return false
		}
		e.triggerCompletion()
		return true
	}

	return false
}

func (e *Editor) triggerCompletion() {
	context, err := e.getAutocompleteContext()
	if err != nil {
		components.ShowError(e.pages, e.app, err)
		return
	}

	line, col := e.sqlInput.CursorPosition()
	request := autocomplete.Request{
		SQL: e.sqlInput.GetText(),
		Position: autocomplete.Position{
			Line:   line,
			Column: col,
		},
		Dialect: e.sqlInput.GetDialect(),
		Context: context,
	}

	result, err := autocomplete.Complete(request)
	if err != nil {
		components.ShowError(e.pages, e.app, err)
		return
	}

	if len(result.Items) == 0 {
		e.hideCompletion()
		return
	}

	e.completionState = completionState{
		items:    result.Items,
		selected: 0,
		replace:  result.Replace,
		active:   true,
	}
	e.completionOverlay.Show(result.Items, 0)
	e.helpBar.SetContext("editor_completion")
	e.app.SetFocus(e.sqlInput)
}

func (e *Editor) hideCompletion() {
	if !e.completionState.active {
		return
	}
	e.completionState.active = false
	e.completionOverlay.Hide()
	e.helpBar.SetContext("editor")
}

func (e *Editor) selectNextCompletion() {
	if !e.completionState.active || len(e.completionState.items) == 0 {
		return
	}
	e.completionState.selected = (e.completionState.selected + 1) % len(e.completionState.items)
	e.completionOverlay.SetSelected(e.completionState.selected)
}

func (e *Editor) selectPrevCompletion() {
	if !e.completionState.active || len(e.completionState.items) == 0 {
		return
	}
	e.completionState.selected--
	if e.completionState.selected < 0 {
		e.completionState.selected = len(e.completionState.items) - 1
	}
	e.completionOverlay.SetSelected(e.completionState.selected)
}

func (e *Editor) applyCompletion() {
	if !e.completionState.active || len(e.completionState.items) == 0 {
		return
	}
	item := e.completionState.items[e.completionState.selected]

	startOffset := e.sqlInput.OffsetAt(e.completionState.replace.Start.Line, e.completionState.replace.Start.Column)
	endOffset := e.sqlInput.OffsetAt(e.completionState.replace.End.Line, e.completionState.replace.End.Column)
	if startOffset > endOffset {
		startOffset, endOffset = endOffset, startOffset
	}

	e.sqlInput.ReplaceRange(startOffset, endOffset, item.Label)
	e.hideCompletion()
}

func (e *Editor) getAutocompleteContext() (autocomplete.Context, error) {
	if e.dbApp == nil || e.dbApp.Explorer == nil {
		return autocomplete.Context{}, nil
	}

	if !e.completionCache.loadedAt.IsZero() && time.Since(e.completionCache.loadedAt) < completionCacheTTL {
		return e.completionCache.context, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutAutocomplete)
	defer cancel()

	schemas, err := e.dbApp.Explorer.GetSchemas(ctx)
	if err != nil {
		return autocomplete.Context{}, err
	}

	if len(schemas) == 0 {
		schemas = []models.Schema{{Name: ""}}
	}

	if e.dbApp.Config != nil && !e.dbApp.Config.UI.ShowSchemas && len(schemas) > 1 {
		schemas = schemas[:1]
	}

	var tables []models.Table
	for _, schema := range schemas {
		schemaTables, err := e.dbApp.Explorer.GetTables(ctx, schema)
		if err != nil {
			return autocomplete.Context{}, err
		}
		tables = append(tables, schemaTables...)
		if len(tables) >= completionMaxTables {
			tables = tables[:completionMaxTables]
			break
		}
	}

	contextResult := autocomplete.Context{
		Tables: make([]autocomplete.Table, 0, len(tables)),
	}

	for _, table := range tables {
		columns, err := e.dbApp.Explorer.GetTableColumns(ctx, table.Name)
		if err != nil {
			logging.Warn().
				Err(err).
				Str("table", table.Name).
				Msg("Failed to load columns for autocomplete")
			continue
		}

		cols := make([]autocomplete.Column, 0, len(columns.Columns))
		for _, col := range columns.Columns {
			cols = append(cols, autocomplete.Column{Name: col.Name})
		}

		contextResult.Tables = append(contextResult.Tables, autocomplete.Table{
			Name:    table.Name,
			Schema:  table.Schema,
			Columns: cols,
		})
	}

	e.completionCache = completionCache{
		context:  contextResult,
		loadedAt: time.Now(),
	}

	return contextResult, nil
}
