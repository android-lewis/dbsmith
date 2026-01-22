package components

import (
	"github.com/android-lewis/dbsmith/internal/tui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/navidys/tvxwidgets"
	"github.com/rivo/tview"
)

type LoadingOverlay struct {
	*tview.Flex
	spinner     *tvxwidgets.Spinner
	messageView *tview.TextView
	cancelHint  *tview.TextView
	visible     bool
	cancellable bool
}

func NewLoadingOverlay() *LoadingOverlay {
	spinner := tvxwidgets.NewSpinner()
	spinner.SetStyle(tvxwidgets.SpinnerArrows)
	spinner.SetBorder(false)
	spinner.SetBackgroundColor(theme.ThemeColors.BackgroundAlt)
	spinner.SetRect(0, 0, 10, 1)

	messageView := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)
	messageView.SetBackgroundColor(theme.ThemeColors.BackgroundAlt)
	messageView.SetTextColor(theme.ThemeColors.Foreground)
	messageView.SetBorder(false)

	cancelHint := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)
	cancelHint.SetBackgroundColor(theme.ThemeColors.BackgroundAlt)
	cancelHint.SetTextColor(theme.ThemeColors.ForegroundMuted)
	cancelHint.SetBorder(false)
	cancelHint.SetText("Press Esc to cancel")

	content := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(spinner, 1, 0, false).
		AddItem(messageView, 1, 0, false).
		AddItem(cancelHint, 1, 0, false).
		AddItem(nil, 0, 1, false)

	content.SetBackgroundColor(theme.ThemeColors.BackgroundAlt)

	centered := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(content, 60, 0, false).
		AddItem(nil, 0, 1, false)

	centered.SetBackgroundColor(theme.ThemeColors.BackgroundAlt)

	overlay := &LoadingOverlay{
		Flex:        centered,
		spinner:     spinner,
		messageView: messageView,
		cancelHint:  cancelHint,
		visible:     false,
		cancellable: false,
	}

	return overlay
}

func (l *LoadingOverlay) Show(message string, cancellable bool) {
	l.visible = true
	l.cancellable = cancellable
	l.messageView.SetText(message)

	if cancellable {
		l.cancelHint.SetText("Press Esc to cancel")
	} else {
		l.cancelHint.SetText("")
	}

	l.spinner.Reset()
}

func (l *LoadingOverlay) Hide() {
	l.visible = false
	l.spinner.Reset()
}

func (l *LoadingOverlay) Pulse() {
	if l.visible {
		l.spinner.Pulse()
	}
}

func (l *LoadingOverlay) IsVisible() bool {
	return l.visible
}

func (l *LoadingOverlay) IsCancellable() bool {
	return l.cancellable
}

func (l *LoadingOverlay) Draw(screen tcell.Screen) {
	if !l.visible {
		return
	}

	l.Flex.Draw(screen)
}
