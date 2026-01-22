package theme

import "github.com/gdamore/tcell/v2"

var ThemeColors = struct {
	Background      tcell.Color
	BackgroundAlt   tcell.Color
	Foreground      tcell.Color
	ForegroundMuted tcell.Color

	Primary tcell.Color
	Success tcell.Color
	Error   tcell.Color
	Warning tcell.Color
	Info    tcell.Color
	Accent  tcell.Color

	Border        tcell.Color
	BorderFocus   tcell.Color
	Selection     tcell.Color
	SelectionText tcell.Color
}{

	Background:      tcell.NewHexColor(0x0d1117),
	BackgroundAlt:   tcell.NewHexColor(0x161b22),
	Foreground:      tcell.NewHexColor(0xe6edf3),
	ForegroundMuted: tcell.NewHexColor(0x848d97),

	Primary: tcell.NewHexColor(0x2f81f7),
	Success: tcell.NewHexColor(0x3fb950),
	Error:   tcell.NewHexColor(0xf85149),
	Warning: tcell.NewHexColor(0xd29922),
	Info:    tcell.NewHexColor(0x58a6ff),
	Accent:  tcell.NewHexColor(0xa371f7),

	Border:        tcell.NewHexColor(0x30363d),
	BorderFocus:   tcell.NewHexColor(0x2f81f7),
	Selection:     tcell.NewHexColor(0x264f78),
	SelectionText: tcell.NewHexColor(0xe6edf3),
}

type ColorName string

const (
	ColorBackground      ColorName = "background"
	ColorBackgroundAlt   ColorName = "background_alt"
	ColorForeground      ColorName = "foreground"
	ColorForegroundMuted ColorName = "foreground_muted"
	ColorPrimary         ColorName = "primary"
	ColorSuccess         ColorName = "success"
	ColorError           ColorName = "error"
	ColorWarning         ColorName = "warning"
	ColorInfo            ColorName = "info"
	ColorAccent          ColorName = "accent"
	ColorBorder          ColorName = "border"
	ColorBorderFocus     ColorName = "border_focus"
	ColorSelection       ColorName = "selection"
	ColorSelectionText   ColorName = "selection_text"
)

func GetColor(name ColorName) tcell.Color {
	switch name {
	case ColorBackground:
		return ThemeColors.Background
	case ColorBackgroundAlt:
		return ThemeColors.BackgroundAlt
	case ColorForeground:
		return ThemeColors.Foreground
	case ColorForegroundMuted:
		return ThemeColors.ForegroundMuted
	case ColorPrimary:
		return ThemeColors.Primary
	case ColorSuccess:
		return ThemeColors.Success
	case ColorError:
		return ThemeColors.Error
	case ColorWarning:
		return ThemeColors.Warning
	case ColorInfo:
		return ThemeColors.Info
	case ColorAccent:
		return ThemeColors.Accent
	case ColorBorder:
		return ThemeColors.Border
	case ColorBorderFocus:
		return ThemeColors.BorderFocus
	case ColorSelection:
		return ThemeColors.Selection
	case ColorSelectionText:
		return ThemeColors.SelectionText
	default:
		return ThemeColors.Foreground
	}
}
