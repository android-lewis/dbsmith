package theme

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func ApplyStyles() {
	tview.Styles.PrimitiveBackgroundColor = ThemeColors.Background
	tview.Styles.ContrastBackgroundColor = ThemeColors.BackgroundAlt
	tview.Styles.MoreContrastBackgroundColor = ThemeColors.BackgroundAlt

	tview.Styles.BorderColor = ThemeColors.Border
	tview.Styles.GraphicsColor = ThemeColors.Border

	tview.Styles.TitleColor = ThemeColors.Foreground
	tview.Styles.PrimaryTextColor = ThemeColors.Foreground
	tview.Styles.SecondaryTextColor = ThemeColors.ForegroundMuted
	tview.Styles.TertiaryTextColor = ThemeColors.ForegroundMuted

	tview.Styles.InverseTextColor = ThemeColors.Background
	tview.Styles.ContrastSecondaryTextColor = ThemeColors.Info
}

func SetBorderFocused(box *tview.Box) {
	if box == nil {
		return
	}
	box.SetBorderColor(ThemeColors.BorderFocus)
	box.SetTitleColor(ThemeColors.Primary)
}

func SetBorderUnfocused(box *tview.Box) {
	if box == nil {
		return
	}
	box.SetBorderColor(ThemeColors.Border)
	box.SetTitleColor(ThemeColors.ForegroundMuted)
}

func SetBorderActive(primitive tview.Primitive) {
	if box, ok := primitive.(*tview.Box); ok {
		SetBorderFocused(box)
	}
}

func SetBorderInactive(primitive tview.Primitive) {
	if box, ok := primitive.(*tview.Box); ok {
		SetBorderUnfocused(box)
	}
}

func StyleBox(box *tview.Box, title string, focused bool) {
	if box == nil {
		return
	}

	box.SetBackgroundColor(ThemeColors.Background)
	box.SetTitle(title)

	if focused {
		SetBorderFocused(box)
	} else {
		SetBorderUnfocused(box)
	}
}

func GetTextStyle(fg, bg tcell.Color) tcell.Style {
	if fg == tcell.ColorDefault {
		fg = ThemeColors.Foreground
	}
	if bg == tcell.ColorDefault {
		bg = ThemeColors.Background
	}
	return tcell.StyleDefault.Foreground(fg).Background(bg)
}

func GetHighlightStyle() tcell.Style {
	return tcell.StyleDefault.
		Foreground(ThemeColors.SelectionText).
		Background(ThemeColors.Selection)
}

func GetErrorStyle() tcell.Style {
	return tcell.StyleDefault.
		Foreground(ThemeColors.Error).
		Background(ThemeColors.Background)
}

func GetSuccessStyle() tcell.Style {
	return tcell.StyleDefault.
		Foreground(ThemeColors.Success).
		Background(ThemeColors.Background)
}

func GetWarningStyle() tcell.Style {
	return tcell.StyleDefault.
		Foreground(ThemeColors.Warning).
		Background(ThemeColors.Background)
}

func GetInfoStyle() tcell.Style {
	return tcell.StyleDefault.
		Foreground(ThemeColors.Info).
		Background(ThemeColors.Background)
}

func ColorTag(name ColorName, bold bool) string {
	color := GetColor(name)
	hex := color.String()
	if bold {
		return "[" + hex + "::b]"
	}
	return "[" + hex + "]"
}

func ColorTagReset() string {
	return "[-:-:-]"
}
