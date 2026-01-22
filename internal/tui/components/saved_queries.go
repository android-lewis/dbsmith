package components

import (
	"fmt"
	"time"

	"github.com/android-lewis/dbsmith/internal/models"
	"github.com/android-lewis/dbsmith/internal/workspace"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SavedQueriesManager struct {
	pages     *tview.Pages
	app       *tview.Application
	workspace *workspace.Manager

	onQuerySave            func(queryID, queryName string)
	onQueryLoad            func(queryID, queryName, querySQL string)
	getSQLText             func() string
	getCurrentSavedQueryID func() string
}

func NewSavedQueriesManager(pages *tview.Pages, app *tview.Application, workspace *workspace.Manager) *SavedQueriesManager {
	return &SavedQueriesManager{
		pages:     pages,
		app:       app,
		workspace: workspace,
	}
}

func (s *SavedQueriesManager) getExistingQuery() *models.SavedQuery {
	if s.getCurrentSavedQueryID == nil {
		return nil
	}
	existingQueryID := s.getCurrentSavedQueryID()
	if existingQueryID == "" {
		return nil
	}
	for _, q := range s.workspace.ListSavedQueries() {
		if q.ID == existingQueryID {
			return &q
		}
	}
	return nil
}

func (s *SavedQueriesManager) SetCallbacks(
	onQuerySave func(queryID, queryName string),
	onQueryLoad func(queryID, queryName, querySQL string),
	getSQLText func() string,
	getCurrentSavedQueryID func() string,
) {
	s.onQuerySave = onQuerySave
	s.onQueryLoad = onQueryLoad
	s.getSQLText = getSQLText
	s.getCurrentSavedQueryID = getCurrentSavedQueryID
}

func (s *SavedQueriesManager) QuickSave(focusWidget tview.Primitive) {
	if s.workspace == nil {
		ShowError(s.pages, s.app, fmt.Errorf("no workspace loaded"))
		return
	}

	workspacePath := workspace.GetDefaultWorkspacePath()
	if workspacePath == "" {
		ShowError(s.pages, s.app, fmt.Errorf("workspace file path not found"))
		return
	}

	existingQuery := s.getExistingQuery()
	if existingQuery == nil {
		s.ShowSaveDialog(focusWidget)
		return
	}

	var sql string
	if s.getSQLText != nil {
		sql = s.getSQLText()
	}

	if sql == "" {
		ShowError(s.pages, s.app, fmt.Errorf("no SQL query to save"))
		return
	}

	updatedQuery := models.SavedQuery{
		ID:          existingQuery.ID,
		Name:        existingQuery.Name,
		SQL:         sql,
		Description: existingQuery.Description,
		CreatedAt:   existingQuery.CreatedAt,
	}

	if err := s.workspace.UpdateSavedQuery(updatedQuery); err != nil {
		ShowError(s.pages, s.app, fmt.Errorf("failed to save query: %w", err))
		return
	}

	if err := s.workspace.Save(workspacePath); err != nil {
		ShowError(s.pages, s.app, fmt.Errorf("failed to save workspace: %w", err))
		return
	}

	if s.onQuerySave != nil {
		s.onQuerySave(existingQuery.ID, existingQuery.Name)
	}

	ShowInfo(s.pages, s.app, fmt.Sprintf("Query saved: %s", existingQuery.Name))
}

func (s *SavedQueriesManager) ShowSaveDialog(focusWidget tview.Primitive) {
	if s.workspace == nil {
		ShowError(s.pages, s.app, fmt.Errorf("no workspace loaded"))
		return
	}

	var sql string
	if s.getSQLText != nil {
		sql = s.getSQLText()
	}

	if sql == "" {
		ShowError(s.pages, s.app, fmt.Errorf("no SQL query to save"))
		return
	}

	existingQuery := s.getExistingQuery()
	isEditMode := existingQuery != nil

	var form *tview.Form
	name := ""
	description := ""

	if isEditMode {
		name = existingQuery.Name
		description = existingQuery.Description
	}

	form = tview.NewForm()

	form.AddInputField("Name", name, 40, nil, func(text string) { name = text })
	form.AddInputField("Description", description, 50, nil, func(text string) { description = text })

	form.AddButton("Save", func() {
		if name == "" {
			ShowError(s.pages, s.app, fmt.Errorf("query name is required"))
			return
		}

		var queryID string
		var err error

		if isEditMode {
			queryID = existingQuery.ID
			updatedQuery := models.SavedQuery{
				ID:          queryID,
				Name:        name,
				SQL:         sql,
				Description: description,
				CreatedAt:   existingQuery.CreatedAt,
			}
			err = s.workspace.UpdateSavedQuery(updatedQuery)
		} else {
			queryID = fmt.Sprintf("query_%d", time.Now().Unix())
			savedQuery := models.SavedQuery{
				ID:          queryID,
				Name:        name,
				SQL:         sql,
				Description: description,
				CreatedAt:   time.Now(),
			}
			err = s.workspace.AddSavedQuery(savedQuery)
		}

		if err != nil {
			ShowError(s.pages, s.app, fmt.Errorf("failed to save query: %w", err))
			return
		}

		workspacePath := workspace.GetDefaultWorkspacePath()
		if workspacePath == "" {
			ShowError(s.pages, s.app, fmt.Errorf("workspace file path not found"))
			return
		}

		if err := s.workspace.Save(workspacePath); err != nil {
			ShowError(s.pages, s.app, fmt.Errorf("failed to save workspace: %w", err))
			return
		}

		s.pages.RemovePage("save-query")
		s.app.SetFocus(focusWidget)

		if s.onQuerySave != nil {
			s.onQuerySave(queryID, name)
		}

		if isEditMode {
			ShowInfo(s.pages, s.app, fmt.Sprintf("Query updated: %s", name))
		} else {
			ShowInfo(s.pages, s.app, fmt.Sprintf("Query saved: %s", name))
		}
	})

	form.AddButton("Cancel", func() {
		s.pages.RemovePage("save-query")
		s.app.SetFocus(focusWidget)
	})

	title := " Save Query "
	if isEditMode {
		title = " Update Query "
	}

	form.SetBorder(true).
		SetTitle(title).
		SetTitleAlign(tview.AlignCenter)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			s.pages.RemovePage("save-query")
			s.app.SetFocus(focusWidget)
			return nil
		}
		return event
	})

	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 13, 0, true).
			AddItem(nil, 0, 1, false), 70, 0, true).
		AddItem(nil, 0, 1, false)

	s.pages.AddPage("save-query", flex, true, true)
	s.app.SetFocus(form)
}

func (s *SavedQueriesManager) ShowLoadDialog(focusWidget tview.Primitive, loadSQL func(sql string)) {
	if s.workspace == nil {
		ShowError(s.pages, s.app, fmt.Errorf("no workspace loaded"))
		return
	}

	savedQueries := s.workspace.ListSavedQueries()
	if len(savedQueries) == 0 {
		ShowInfo(s.pages, s.app, "No saved queries found")
		return
	}

	list := tview.NewList().
		ShowSecondaryText(true).
		SetHighlightFullLine(true)

	list.SetBorder(true).
		SetTitle(" Load Saved Query (Enter to load, Esc to cancel) ").
		SetTitleAlign(tview.AlignCenter)

	for _, query := range savedQueries {
		q := query
		secondaryText := query.Description
		if secondaryText == "" {
			secondaryText = fmt.Sprintf("Created: %s", query.CreatedAt.Format("2006-01-02 15:04"))
		}
		list.AddItem(query.Name, secondaryText, 0, func() {
			if loadSQL != nil {
				loadSQL(q.SQL)
			}

			if s.onQueryLoad != nil {
				s.onQueryLoad(q.ID, q.Name, q.SQL)
			}

			s.pages.RemovePage("load-query")
			s.app.SetFocus(focusWidget)
			ShowInfo(s.pages, s.app, fmt.Sprintf("Loaded query: %s", q.Name))
		})
	}

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			s.pages.RemovePage("load-query")
			s.app.SetFocus(focusWidget)
			return nil
		}
		return event
	})

	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(list, 15, 0, true).
			AddItem(nil, 0, 1, false), 60, 0, true).
		AddItem(nil, 0, 1, false)

	s.pages.AddPage("load-query", flex, true, true)
	s.app.SetFocus(list)
}
