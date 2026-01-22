package components

import (
	"fmt"

	"github.com/android-lewis/dbsmith/internal/app"
	"github.com/android-lewis/dbsmith/internal/tui/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/navidys/tvxwidgets"
	"github.com/rivo/tview"
)

type StatusBar struct {
	*tview.Flex
	app        *app.App
	textView   *tview.TextView
	spinner    *tvxwidgets.Spinner
	loading    bool
	loadingMsg string
}

func NewStatusBar(application *app.App) *StatusBar {
	spinner := tvxwidgets.NewSpinner()
	spinner.SetStyle(tvxwidgets.SpinnerDotsCircling)
	spinner.SetBorder(false)
	spinner.SetBackgroundColor(theme.ThemeColors.BackgroundAlt)

	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	tv.SetBackgroundColor(theme.ThemeColors.BackgroundAlt)
	tv.SetTextColor(theme.ThemeColors.Foreground)

	flex := tview.NewFlex().
		SetDirection(tview.FlexColumn)
	flex.SetBackgroundColor(theme.ThemeColors.BackgroundAlt)
	flex.AddItem(tv, 0, 1, false)

	sb := &StatusBar{
		Flex:     flex,
		app:      application,
		textView: tv,
		spinner:  spinner,
		loading:  false,
	}

	sb.Update()
	return sb
}

func (s *StatusBar) Update() {
	var text string

	if s.loading {
		text = fmt.Sprintf(" [#%06x::b]%s[-:-:-] %s",
			theme.ThemeColors.Info.Hex(),
			theme.Icons.Loading,
			s.loadingMsg)
	} else if s.app == nil || s.app.Connection == nil {
		text = fmt.Sprintf(" [#%06x::b]%s[-:-:-] No Connection",
			theme.ThemeColors.Warning.Hex(),
			theme.Icons.Warning)
	} else {
		conn := s.app.Connection
		var statusIcon string
		var statusColor tcell.Color

		if s.app.Driver == nil {
			statusIcon = theme.Icons.Disconnected
			statusColor = theme.ThemeColors.Error
		} else {
			statusIcon = theme.Icons.Connected
			statusColor = theme.ThemeColors.Success
		}

		text = fmt.Sprintf(" [#%06x::b]%s[-:-:-] [#%06x::b]%s[-:-:-] [#%06x]│[-] %s [#%06x]│[-] %s@%s:%d/%s",
			statusColor.Hex(),
			statusIcon,
			theme.ThemeColors.Foreground.Hex(),
			conn.Name,
			theme.ThemeColors.ForegroundMuted.Hex(),
			conn.Type,
			theme.ThemeColors.ForegroundMuted.Hex(),
			conn.Username,
			conn.Host,
			conn.Port,
			conn.Database,
		)
	}

	s.textView.SetText(text)
}

func (s *StatusBar) SetLoading(message string) {
	s.loading = true
	s.loadingMsg = message

	if s.GetItemCount() == 1 {
		s.AddItem(s.spinner, 5, 0, false)
	}

	s.Update()
}

func (s *StatusBar) SetIdle() {
	s.loading = false
	s.loadingMsg = ""

	if s.GetItemCount() > 1 {
		s.RemoveItem(s.spinner)
	}

	s.Update()
}
