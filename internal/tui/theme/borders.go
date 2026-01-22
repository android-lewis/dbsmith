package theme

import "github.com/rivo/tview"

const (
	roundedTopLeft     = '╭'
	roundedTopRight    = '╮'
	roundedBottomLeft  = '╰'
	roundedBottomRight = '╯'
	roundedHorizontal  = '─'
	roundedVertical    = '│'

	squareTopLeft     = '+'
	squareTopRight    = '+'
	squareBottomLeft  = '+'
	squareBottomRight = '+'
	squareHorizontal  = '-'
	squareVertical    = '|'
)

func ApplyRoundedBorders() {
	tview.Borders.HorizontalFocus = roundedHorizontal
	tview.Borders.VerticalFocus = roundedVertical
	tview.Borders.TopLeftFocus = roundedTopLeft
	tview.Borders.TopRightFocus = roundedTopRight
	tview.Borders.BottomLeftFocus = roundedBottomLeft
	tview.Borders.BottomRightFocus = roundedBottomRight

	tview.Borders.Horizontal = roundedHorizontal
	tview.Borders.Vertical = roundedVertical
	tview.Borders.TopLeft = roundedTopLeft
	tview.Borders.TopRight = roundedTopRight
	tview.Borders.BottomLeft = roundedBottomLeft
	tview.Borders.BottomRight = roundedBottomRight
}

func ApplySquareBorders() {
	tview.Borders.HorizontalFocus = squareHorizontal
	tview.Borders.VerticalFocus = squareVertical
	tview.Borders.TopLeftFocus = squareTopLeft
	tview.Borders.TopRightFocus = squareTopRight
	tview.Borders.BottomLeftFocus = squareBottomLeft
	tview.Borders.BottomRightFocus = squareBottomRight

	tview.Borders.Horizontal = squareHorizontal
	tview.Borders.Vertical = squareVertical
	tview.Borders.TopLeft = squareTopLeft
	tview.Borders.TopRight = squareTopRight
	tview.Borders.BottomLeft = squareBottomLeft
	tview.Borders.BottomRight = squareBottomRight
}
