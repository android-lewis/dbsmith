package components

import (
	"fmt"
	"time"

	"github.com/android-lewis/dbsmith/internal/tui/theme"
	"github.com/android-lewis/dbsmith/internal/tui/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/navidys/tvxwidgets"
	"github.com/rivo/tview"
)

type QueryStats struct {
	*tview.Flex
	sparkline     *tvxwidgets.Sparkline
	statsView     *tview.TextView
	queryTimes    []float64
	lastDuration  time.Duration
	lastRowCount  int
	maxDataPoints int
}

func NewQueryStats() *QueryStats {
	maxDataPoints := 10

	sparkline := tvxwidgets.NewSparkline()
	sparkline.SetBorder(false)
	sparkline.SetBackgroundColor(theme.ThemeColors.Background)
	sparkline.SetLineColor(theme.ThemeColors.Primary)
	sparkline.SetDataTitle("")
	sparkline.SetRect(0, 0, 20, 3)

	statsView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	statsView.SetBackgroundColor(theme.ThemeColors.Background)
	statsView.SetTextColor(theme.ThemeColors.Foreground)
	statsView.SetBorder(false)

	flex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(sparkline, 20, 0, false).
		AddItem(statsView, 0, 1, false)

	flex.SetBackgroundColor(theme.ThemeColors.Background)
	flex.SetBorder(true)
	flex.SetTitle(" Query Stats ")
	flex.SetTitleColor(theme.ThemeColors.Primary)
	flex.SetBorderColor(theme.ThemeColors.Border)

	qs := &QueryStats{
		Flex:          flex,
		sparkline:     sparkline,
		statsView:     statsView,
		queryTimes:    make([]float64, 0, maxDataPoints),
		maxDataPoints: maxDataPoints,
	}

	qs.updateDisplay()
	return qs
}

func (q *QueryStats) RecordQuery(duration time.Duration, rowCount int) {
	q.lastDuration = duration
	q.lastRowCount = rowCount

	durationMs := float64(duration.Milliseconds())

	q.queryTimes = append(q.queryTimes, durationMs)
	if len(q.queryTimes) > q.maxDataPoints {
		q.queryTimes = q.queryTimes[1:]
	}

	q.sparkline.SetData(q.queryTimes)

	q.updateDisplay()
}

func (q *QueryStats) updateDisplay() {
	if q.lastDuration == 0 {
		q.statsView.SetText(" No queries executed yet")
		return
	}

	var durationStr string
	if q.lastDuration < time.Second {
		durationStr = fmt.Sprintf("%.3fs", q.lastDuration.Seconds())
	} else {
		durationStr = fmt.Sprintf("%.2fs", q.lastDuration.Seconds())
	}

	rowCountStr := utils.FormatNumber(int64(q.lastRowCount))

	var avgStr string
	if len(q.queryTimes) > 0 {
		var sum float64
		for _, t := range q.queryTimes {
			sum += t
		}
		avg := sum / float64(len(q.queryTimes))
		avgStr = fmt.Sprintf("%.0fms", avg)
	} else {
		avgStr = "N/A"
	}

	text := fmt.Sprintf(" [#%06x::b]Last:[-:-:-] %s\n [#%06x::b]Rows:[-:-:-] %s\n [#%06x::b]Avg:[-:-:-]  %s",
		theme.ThemeColors.Info.Hex(),
		durationStr,
		theme.ThemeColors.Success.Hex(),
		rowCountStr,
		theme.ThemeColors.Accent.Hex(),
		avgStr,
	)

	q.statsView.SetText(text)
}

func (q *QueryStats) Reset() {
	q.queryTimes = make([]float64, 0, q.maxDataPoints)
	q.lastDuration = 0
	q.lastRowCount = 0
	q.sparkline.SetData([]float64{})
	q.updateDisplay()
}

func (q *QueryStats) GetLastDuration() time.Duration {
	return q.lastDuration
}

func (q *QueryStats) GetLastRowCount() int {
	return q.lastRowCount
}

func (q *QueryStats) GetQueryCount() int {
	return len(q.queryTimes)
}

func (q *QueryStats) Draw(screen tcell.Screen) {
	q.Flex.Draw(screen)
}
