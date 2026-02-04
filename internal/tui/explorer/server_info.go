package explorer

import (
	"context"
	"fmt"

	"github.com/android-lewis/dbsmith/internal/tui/constants"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showServerInfo displays a modal with database server information
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
