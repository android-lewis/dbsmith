package theme

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Theme struct {
	ColorPrimary   tcell.Color
	ColorSecondary tcell.Color

	ColorSuccess tcell.Color
	ColorError   tcell.Color
	ColorWarning tcell.Color

	ColorBackground tcell.Color
	ColorText       tcell.Color

	ColorBorder      tcell.Color
	ColorBorderFocus tcell.Color

	ColorTitle  tcell.Color
	ColorSubtle tcell.Color
}

func Default() *Theme {
	return &Theme{
		ColorPrimary:     ThemeColors.Primary,
		ColorSecondary:   ThemeColors.Info,
		ColorSuccess:     ThemeColors.Success,
		ColorError:       ThemeColors.Error,
		ColorWarning:     ThemeColors.Warning,
		ColorBackground:  ThemeColors.Background,
		ColorText:        ThemeColors.Foreground,
		ColorBorder:      ThemeColors.Border,
		ColorBorderFocus: ThemeColors.BorderFocus,
		ColorTitle:       ThemeColors.Foreground,
		ColorSubtle:      ThemeColors.ForegroundMuted,
	}
}

func (t *Theme) Apply() {
	tview.Styles.PrimitiveBackgroundColor = t.ColorBackground
	tview.Styles.ContrastBackgroundColor = ThemeColors.BackgroundAlt
	tview.Styles.MoreContrastBackgroundColor = ThemeColors.BackgroundAlt
	tview.Styles.BorderColor = t.ColorBorder
	tview.Styles.TitleColor = t.ColorTitle
	tview.Styles.GraphicsColor = t.ColorBorder
	tview.Styles.PrimaryTextColor = t.ColorText
	tview.Styles.SecondaryTextColor = t.ColorSubtle
	tview.Styles.TertiaryTextColor = t.ColorSubtle
	tview.Styles.InverseTextColor = ThemeColors.Background
	tview.Styles.ContrastSecondaryTextColor = ThemeColors.Info
}

func ApplyTheme() {
	ApplyStyles()
	ApplyRoundedBorders()
}
