package components

import (
	"strings"

	"github.com/android-lewis/dbsmith/internal/tui/syntax"
	"github.com/android-lewis/dbsmith/internal/tui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SQLEditor struct {
	*tview.TextArea

	highlighter    *syntax.Highlighter
	dialect        string
	styledTokens   []syntax.StyledToken
	contentChanged bool

	userChangedFunc func()
}

func NewSQLEditor() *SQLEditor {
	textArea := tview.NewTextArea()
	textArea.SetBorder(true)
	textArea.SetBackgroundColor(theme.ThemeColors.Background)
	textArea.SetTextStyle(tcell.StyleDefault.
		Foreground(theme.ThemeColors.Foreground).
		Background(theme.ThemeColors.Background))
	textArea.SetSelectedStyle(tcell.StyleDefault.
		Foreground(theme.ThemeColors.SelectionText).
		Background(theme.ThemeColors.Selection))

	e := &SQLEditor{
		TextArea:       textArea,
		highlighter:    syntax.NewHighlighter(syntax.DefaultTheme()),
		dialect:        "sql",
		contentChanged: true,
	}

	textArea.SetChangedFunc(func() {
		e.contentChanged = true
		if e.userChangedFunc != nil {
			e.userChangedFunc()
		}
	})

	return e
}

func (e *SQLEditor) GetText() string {
	return e.TextArea.GetText()
}

func (e *SQLEditor) SetText(text string, cursorAtEnd bool) *SQLEditor {
	e.TextArea.SetText(text, cursorAtEnd)
	e.contentChanged = true
	return e
}

func (e *SQLEditor) SetChangedFunc(handler func()) *SQLEditor {
	e.userChangedFunc = handler
	return e
}

func (e *SQLEditor) SetHighlighter(h *syntax.Highlighter) *SQLEditor {
	e.highlighter = h
	e.contentChanged = true
	return e
}

func (e *SQLEditor) SetDialect(dialect string) *SQLEditor {
	e.dialect = dialect
	e.contentChanged = true
	return e
}

func (e *SQLEditor) GetDialect() string {
	return e.dialect
}

func (e *SQLEditor) SetPlaceholder(placeholder string) *SQLEditor {

	lines := strings.Split(placeholder, "\n")
	if len(lines) > 0 {
		e.TextArea.SetPlaceholder(lines[0])
	}
	return e
}

func (e *SQLEditor) SetInputCapture(capture func(event *tcell.EventKey) *tcell.EventKey) *SQLEditor {
	e.TextArea.SetInputCapture(capture)
	return e
}

func (e *SQLEditor) CursorPosition() (int, int) {
	_, _, row, col := e.TextArea.GetCursor()
	return row, col
}

func (e *SQLEditor) OffsetAt(line, col int) int {
	lines := strings.Split(e.GetText(), "\n")
	return e.offsetAt(lines, line, col)
}

func (e *SQLEditor) ReplaceRange(start, end int, text string) {
	e.TextArea.Replace(start, end, text)
	e.contentChanged = true
}

func (e *SQLEditor) Draw(screen tcell.Screen) {

	e.TextArea.Draw(screen)

	if e.highlighter == nil {
		return
	}

	if e.contentChanged {
		tokens, _ := e.highlighter.Highlight(e.GetText(), e.dialect)
		e.styledTokens = tokens
		e.contentChanged = false
	}

	x, y, width, height := e.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}

	rowOffset, colOffset := e.GetOffset()

	tokenMap := make(map[int]map[int]tcell.Style)
	for _, token := range e.styledTokens {
		if tokenMap[token.Line] == nil {
			tokenMap[token.Line] = make(map[int]tcell.Style)
		}
		for i := 0; i < len(token.Text); i++ {
			tokenMap[token.Line][token.Col+i] = token.Style
		}
	}

	text := e.GetText()
	lines := strings.Split(text, "\n")

	if len(lines) == 1 && lines[0] == "" {
		return
	}

	for row := 0; row < height && rowOffset+row < len(lines); row++ {
		lineIdx := rowOffset + row
		line := lines[lineIdx]
		lineRunes := []rune(line)

		for col := 0; col < width && colOffset+col < len(lineRunes); col++ {
			charIdx := colOffset + col
			if charIdx >= len(lineRunes) {
				break
			}

			if lineStyles, ok := tokenMap[lineIdx]; ok {
				if style, ok := lineStyles[charIdx]; ok {

					ch := lineRunes[charIdx]

					_, start, end := e.GetSelection()
					offset := e.offsetAt(lines, lineIdx, charIdx)
					if start != end && offset >= start && offset < end {

						continue
					}
					screen.SetContent(x+col, y+row, ch, nil, style)
				}
			}
		}
	}
}

func (e *SQLEditor) offsetAt(lines []string, line, col int) int {
	offset := 0
	for i := 0; i < line && i < len(lines); i++ {
		offset += len(lines[i]) + 1
	}
	if line < len(lines) && col <= len(lines[line]) {
		offset += col
	}
	return offset
}
