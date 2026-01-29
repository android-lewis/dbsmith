package components

import (
	"fmt"
	"strings"

	"github.com/android-lewis/dbsmith/internal/tui/theme"
	"github.com/rivo/tview"
)

type HelpBar struct {
	*tview.TextView
	context  string
	expanded bool
}

type KeyHelp struct {
	Key  string
	Desc string
}

var helpContexts = map[string][]KeyHelp{
	"workspace": {
		{Key: "Enter", Desc: "Select"},
		{Key: "N", Desc: "New"},
		{Key: "E", Desc: "Edit"},
		{Key: "D", Desc: "Delete"},
		{Key: "T", Desc: "Test"},
		{Key: "F10", Desc: "Quit"},
		{Key: "F1", Desc: "More"},
	},
	"explorer": {
		{Key: "Tab", Desc: "Panels"},
		{Key: "Alt+H", Desc: "Schemas"},
		{Key: "Alt+I", Desc: "Indexes"},
		{Key: "Alt+D", Desc: "Data"},
		{Key: "F3", Desc: "Editor"},
		{Key: "F10", Desc: "Quit"},
		{Key: "F1", Desc: "More"},
	},
	"editor": {
		{Key: "F5", Desc: "Run"},
		{Key: "Esc", Desc: "Cancel"},
		{Key: "Tab", Desc: "Complete"},
		{Key: "Alt+S", Desc: "Save"},
		{Key: "Alt+Shift+S", Desc: "Save As"},
		{Key: "Alt+L", Desc: "Load"},
		{Key: "Alt+T", Desc: "NewTab"},
		{Key: "F1", Desc: "More"},
	},
	"editor_completion": {
		{Key: "Tab", Desc: "next"},
		{Key: "Shift+Tab", Desc: "prev"},
		{Key: "Enter", Desc: "insert"},
		{Key: "Esc", Desc: "cancel"},
		{Key: "F5", Desc: "Run"},
		{Key: "Alt+S", Desc: "Save"},
		{Key: "F1", Desc: "More"},
	},
}

var helpContextsExpanded = map[string][]KeyHelp{
	"workspace": {
		{Key: "Enter", Desc: "Select/connect to connection"},
		{Key: "N", Desc: "New connection"},
		{Key: "E", Desc: "Edit connection"},
		{Key: "D", Desc: "Delete connection"},
		{Key: "T", Desc: "Test connection"},
		{Key: "Esc", Desc: "Cancel / Back"},
		{Key: "F1", Desc: "Collapse help"},
	},
	"explorer": {
		{Key: "Tab", Desc: "Cycle between panels"},
		{Key: "Alt+H", Desc: "Toggle schemas panel"},
		{Key: "Alt+I", Desc: "Toggle indexes panel"},
		{Key: "Alt+D", Desc: "Toggle data preview"},
		{Key: "Enter", Desc: "Select item"},
		{Key: "PgUp/PgDn", Desc: "Scroll data preview"},
		{Key: "F1", Desc: "Collapse help"},
	},
	"editor": {
		{Key: "F5/Shift+Enter", Desc: "Execute query"},
		{Key: "Esc", Desc: "Cancel running query"},
		{Key: "Alt+M", Desc: "Toggle Execute/Analyze"},
		{Key: "Alt+S", Desc: "Quick save"},
		{Key: "Alt+Shift+S", Desc: "Save As"},
		{Key: "Alt+L", Desc: "Load saved query"},
		{Key: "Alt+E", Desc: "Export results"},
		{Key: "Alt+T", Desc: "New tab"},
		{Key: "Alt+W", Desc: "Close tab"},
		{Key: "Alt+R", Desc: "Rename tab"},
		{Key: "Alt+1-9", Desc: "Switch to tab"},
		{Key: "F1", Desc: "Collapse help"},
	},
	"editor_completion": {
		{Key: "Tab", Desc: "Next suggestion"},
		{Key: "Shift+Tab", Desc: "Previous suggestion"},
		{Key: "Enter", Desc: "Insert suggestion"},
		{Key: "Esc", Desc: "Cancel completion"},
		{Key: "F5/Shift+Enter", Desc: "Execute query"},
		{Key: "Alt+S", Desc: "Quick save"},
		{Key: "Alt+L", Desc: "Load saved query"},
		{Key: "Alt+E", Desc: "Export results"},
		{Key: "F1", Desc: "Collapse help"},
	},
}

func NewHelpBar() *HelpBar {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	tv.SetBackgroundColor(theme.ThemeColors.BackgroundAlt)
	tv.SetTextColor(theme.ThemeColors.Foreground)

	hb := &HelpBar{
		TextView: tv,
		context:  "workspace",
	}

	hb.update()
	return hb
}

func (h *HelpBar) SetContext(context string) {
	h.context = context
	h.update()
}

func (h *HelpBar) ToggleExpanded() {
	h.expanded = !h.expanded
	h.update()
}

func (h *HelpBar) IsExpanded() bool {
	return h.expanded
}

func (h *HelpBar) GetHeight() int {
	if h.expanded {
		helps := helpContextsExpanded[h.context]
		if helps == nil {
			helps = helpContextsExpanded["workspace"]
		}
		itemsPerRow := 4
		rows := (len(helps) + itemsPerRow - 1) / itemsPerRow
		if rows < 2 {
			rows = 2
		}
		return rows
	}
	return 1
}

func (h *HelpBar) update() {
	if h.expanded {
		h.updateExpanded()
	} else {
		h.updateCollapsed()
	}
}

func (h *HelpBar) updateCollapsed() {
	helps, ok := helpContexts[h.context]
	if !ok {
		helps = helpContexts["workspace"]
	}

	var parts []string
	for _, help := range helps {
		parts = append(parts, fmt.Sprintf("[#%06x::b]%s[-:-:-] [#%06x]%s[-]",
			theme.ThemeColors.Primary.Hex(), help.Key,
			theme.ThemeColors.ForegroundMuted.Hex(), help.Desc))
	}

	separator := fmt.Sprintf("[#%06x]│[-]", theme.ThemeColors.ForegroundMuted.Hex())
	leftContent := " " + strings.Join(parts, "  "+separator+"  ")

	contextName := h.context
	if len(contextName) > 0 {
		contextName = strings.ToUpper(contextName[:1]) + contextName[1:]
	}
	rightContent := fmt.Sprintf("[#%06x][[#%06x]%s[#%06x]][-]",
		theme.ThemeColors.ForegroundMuted.Hex(),
		theme.ThemeColors.Info.Hex(),
		contextName,
		theme.ThemeColors.ForegroundMuted.Hex())

	_, _, width, _ := h.GetInnerRect()
	if width > 0 {
		leftLen := stripColorCodes(leftContent)
		rightLen := stripColorCodes(rightContent)
		padding := width - leftLen - rightLen - 2
		if padding > 0 {
			h.SetText(leftContent + strings.Repeat(" ", padding) + rightContent + " ")
		} else {
			h.SetText(leftContent + " " + rightContent + " ")
		}
	} else {
		h.SetText(leftContent + " " + rightContent + " ")
	}
}

func (h *HelpBar) updateExpanded() {
	helps, ok := helpContextsExpanded[h.context]
	if !ok {
		helps = helpContextsExpanded["workspace"]
	}

	contextName := h.context
	if len(contextName) > 0 {
		contextName = strings.ToUpper(contextName[:1]) + contextName[1:]
	}

	_, _, width, _ := h.GetInnerRect()
	if width <= 0 {
		width = 80
	}

	itemsPerRow := 4
	separator := fmt.Sprintf("[#%06x]│[-]", theme.ThemeColors.ForegroundMuted.Hex())

	var lines []string
	for i := 0; i < len(helps); i += itemsPerRow {
		end := i + itemsPerRow
		if end > len(helps) {
			end = len(helps)
		}

		var rowParts []string
		for _, help := range helps[i:end] {
			rowParts = append(rowParts, fmt.Sprintf("[#%06x::b]%s[-:-:-] [#%06x]%s[-]",
				theme.ThemeColors.Primary.Hex(), help.Key,
				theme.ThemeColors.ForegroundMuted.Hex(), help.Desc))
		}

		rowContent := " " + strings.Join(rowParts, "  "+separator+"  ")
		lines = append(lines, rowContent)
	}

	if len(lines) > 0 {
		lastIdx := len(lines) - 1
		rightContent := fmt.Sprintf("[#%06x][[#%06x]%s[#%06x]][-]",
			theme.ThemeColors.ForegroundMuted.Hex(),
			theme.ThemeColors.Info.Hex(),
			contextName,
			theme.ThemeColors.ForegroundMuted.Hex())

		leftLen := stripColorCodes(lines[lastIdx])
		rightLen := stripColorCodes(rightContent)
		padding := width - leftLen - rightLen - 2
		if padding > 0 {
			lines[lastIdx] = lines[lastIdx] + strings.Repeat(" ", padding) + rightContent + " "
		} else {
			lines[lastIdx] = lines[lastIdx] + " " + rightContent + " "
		}
	}

	h.SetText(strings.Join(lines, "\n"))
}

func stripColorCodes(s string) int {
	inTag := false
	length := 0
	for _, r := range s {
		if r == '[' {
			inTag = true
		} else if r == ']' {
			inTag = false
		} else if !inTag {
			length++
		}
	}
	return length
}
