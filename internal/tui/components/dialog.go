package components

import (
	"github.com/rivo/tview"
)

func ShowError(pages *tview.Pages, app *tview.Application, err error) {
	dialog := tview.NewModal().
		SetText("Error: " + err.Error()).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("error")
		})

	pages.AddPage("error", dialog, true, true)
	app.SetFocus(dialog)
}

func ShowInfo(pages *tview.Pages, app *tview.Application, message string) {
	dialog := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("info")
		})

	pages.AddPage("info", dialog, true, true)
	app.SetFocus(dialog)
}

func ShowConfirm(pages *tview.Pages, app *tview.Application, message string, callback func(bool)) {
	dialog := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("confirm")
			callback(buttonIndex == 0)
		})

	pages.AddPage("confirm", dialog, true, true)
	app.SetFocus(dialog)
}
