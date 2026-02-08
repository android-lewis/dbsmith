package editor

import (
	"context"
	"strings"
	"sync"
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
	tables        []models.Table
	tablesLoaded  time.Time
	columnsByName map[string]columnCacheEntry
}

type columnCacheEntry struct {
	columns []autocomplete.Column
	loaded  time.Time
}

const completionMaxTables = 200
const completionTablesTTL = 30 * time.Second
const completionColumnsTTL = 2 * time.Minute
const completionColumnWorkers = 4

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
	line, col := e.sqlInput.CursorPosition()
	request := autocomplete.Request{
		SQL: e.sqlInput.GetText(),
		Position: autocomplete.Position{
			Line:   line,
			Column: col,
		},
		Dialect: e.sqlInput.GetDialect(),
	}

	analysis, err := autocomplete.Analyze(request)
	if err != nil {
		components.ShowError(e.pages, e.app, err)
		return
	}

	if analysis.Suppress {
		e.hideCompletion()
		return
	}

	context, err := e.getAutocompleteContext(analysis)
	if err != nil {
		components.ShowError(e.pages, e.app, err)
		return
	}

	result := autocomplete.CompleteWithAnalysis(analysis, context, request.Dialect)

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

func (e *Editor) getAutocompleteContext(analysis autocomplete.Analysis) (autocomplete.Context, error) {
	if e.dbApp == nil || e.dbApp.Explorer == nil {
		return autocomplete.Context{}, nil
	}

	needsTables := analysis.HasKind(autocomplete.KindTable) || analysis.HasKind(autocomplete.KindColumn)
	if !needsTables {
		return autocomplete.Context{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutAutocomplete)
	defer cancel()

	tables, err := e.loadCompletionTables(ctx)
	if err != nil {
		return autocomplete.Context{}, err
	}

	if len(tables) > completionMaxTables {
		tables = tables[:completionMaxTables]
	}

	if !analysis.HasKind(autocomplete.KindColumn) {
		return buildContextFromTables(tables, nil), nil
	}

	targets := analysis.TargetTables()
	columnsByTable := e.loadCompletionColumns(ctx, tables, targets)

	return buildContextFromTables(tables, columnsByTable), nil
}

func (e *Editor) loadCompletionTables(ctx context.Context) ([]models.Table, error) {
	if !e.completionCache.tablesLoaded.IsZero() && time.Since(e.completionCache.tablesLoaded) < completionTablesTTL {
		return e.completionCache.tables, nil
	}

	schemas, err := e.dbApp.Explorer.GetSchemas(ctx)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		tables = append(tables, schemaTables...)
		if len(tables) >= completionMaxTables {
			tables = tables[:completionMaxTables]
			break
		}
	}

	e.completionCache.tables = tables
	e.completionCache.tablesLoaded = time.Now()
	return tables, nil
}

func (e *Editor) loadCompletionColumns(ctx context.Context, tables []models.Table, targetNames []string) map[string][]autocomplete.Column {
	if len(tables) == 0 {
		return nil
	}

	if e.completionCache.columnsByName == nil {
		e.completionCache.columnsByName = make(map[string]columnCacheEntry)
	}

	columnsByTable := map[string][]autocomplete.Column{}
	targetSet := map[string]bool{}
	for _, name := range targetNames {
		targetSet[strings.ToUpper(name)] = true
	}

	var tablesToLoad []models.Table
	for _, table := range tables {
		key := strings.ToUpper(table.Name)
		if len(targetSet) > 0 && !targetSet[key] {
			continue
		}
		entry, ok := e.completionCache.columnsByName[key]
		if ok && time.Since(entry.loaded) < completionColumnsTTL {
			columnsByTable[key] = entry.columns
			continue
		}
		tablesToLoad = append(tablesToLoad, table)
	}

	if len(tablesToLoad) == 0 {
		return columnsByTable
	}

	var mu sync.Mutex
	newEntries := make(map[string]columnCacheEntry)
	var wg sync.WaitGroup
	sem := make(chan struct{}, completionColumnWorkers)

	for _, table := range tablesToLoad {
		table := table
		key := strings.ToUpper(table.Name)
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			columns, err := e.dbApp.Explorer.GetTableColumns(ctx, table.Name)
			if err != nil {
				logging.Warn().
					Err(err).
					Str("table", table.Name).
					Msg("Failed to load columns for autocomplete")
				return
			}

			cols := make([]autocomplete.Column, 0, len(columns.Columns))
			for _, col := range columns.Columns {
				cols = append(cols, autocomplete.Column{Name: col.Name})
			}

			mu.Lock()
			columnsByTable[key] = cols
			newEntries[key] = columnCacheEntry{
				columns: cols,
				loaded:  time.Now(),
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	for key, entry := range newEntries {
		e.completionCache.columnsByName[key] = entry
	}

	return columnsByTable
}

func buildContextFromTables(tables []models.Table, columnsByTable map[string][]autocomplete.Column) autocomplete.Context {
	contextResult := autocomplete.Context{
		Tables: make([]autocomplete.Table, 0, len(tables)),
	}

	for _, table := range tables {
		key := strings.ToUpper(table.Name)
		contextResult.Tables = append(contextResult.Tables, autocomplete.Table{
			Name:    table.Name,
			Schema:  table.Schema,
			Columns: columnsByTable[key],
		})
	}

	return contextResult
}
