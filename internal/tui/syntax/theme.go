package syntax

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/gdamore/tcell/v2"
)

type Theme struct {
	Name  string
	Dark  bool
	style map[chroma.TokenType]tcell.Style
}

func DefaultTheme() *Theme {

	base := tcell.StyleDefault.Background(tcell.NewRGBColor(13, 17, 23)).Foreground(tcell.NewRGBColor(230, 237, 243))

	return &Theme{
		Name: "github-dark",
		Dark: true,
		style: map[chroma.TokenType]tcell.Style{

			chroma.Keyword:            base.Foreground(tcell.NewRGBColor(255, 123, 114)),
			chroma.KeywordNamespace:   base.Foreground(tcell.NewRGBColor(255, 123, 114)),
			chroma.KeywordType:        base.Foreground(tcell.NewRGBColor(255, 123, 114)),
			chroma.KeywordDeclaration: base.Foreground(tcell.NewRGBColor(255, 123, 114)),

			chroma.String:         base.Foreground(tcell.NewRGBColor(168, 219, 255)),
			chroma.StringDouble:   base.Foreground(tcell.NewRGBColor(168, 219, 255)),
			chroma.StringSingle:   base.Foreground(tcell.NewRGBColor(168, 219, 255)),
			chroma.StringBacktick: base.Foreground(tcell.NewRGBColor(168, 219, 255)),

			chroma.Number:        base.Foreground(tcell.NewRGBColor(121, 192, 255)),
			chroma.NumberInteger: base.Foreground(tcell.NewRGBColor(121, 192, 255)),
			chroma.NumberFloat:   base.Foreground(tcell.NewRGBColor(121, 192, 255)),

			chroma.Comment:          base.Foreground(tcell.NewRGBColor(139, 148, 158)),
			chroma.CommentSingle:    base.Foreground(tcell.NewRGBColor(139, 148, 158)),
			chroma.CommentMultiline: base.Foreground(tcell.NewRGBColor(139, 148, 158)),

			chroma.Operator:    base.Foreground(tcell.NewRGBColor(255, 123, 114)),
			chroma.Punctuation: base.Foreground(tcell.NewRGBColor(230, 237, 243)),

			chroma.Name:          base.Foreground(tcell.NewRGBColor(230, 237, 243)),
			chroma.NameBuiltin:   base.Foreground(tcell.NewRGBColor(121, 192, 255)),
			chroma.NameFunction:  base.Foreground(tcell.NewRGBColor(210, 153, 255)),
			chroma.NameClass:     base.Foreground(tcell.NewRGBColor(255, 199, 119)),
			chroma.NameNamespace: base.Foreground(tcell.NewRGBColor(255, 199, 119)),

			chroma.Text:  base.Foreground(tcell.NewRGBColor(230, 237, 243)),
			chroma.Other: base.Foreground(tcell.NewRGBColor(230, 237, 243)),
		},
	}
}

func (t *Theme) ChromaToTcell(token chroma.Token) tcell.Style {
	return t.GetStyle(token.Type)
}

func (t *Theme) GetStyle(tokenType chroma.TokenType) tcell.Style {
	if style, ok := t.style[tokenType]; ok {
		return style
	}

	for tokenType != 0 {
		tokenType = tokenType.Parent()
		if style, ok := t.style[tokenType]; ok {
			return style
		}
	}

	return tcell.StyleDefault.Background(tcell.NewRGBColor(13, 17, 23)).Foreground(tcell.NewRGBColor(230, 237, 243))
}
