package views

import (
	"fmt"

	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"

	"github.com/rivo/tview"
)

type TitleBar struct {
	View  *tview.Flex
	left  *tview.TextView
	right *tview.TextView
}

func NewTitleBar() *TitleBar {
	left := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	left.SetBackgroundColor(theme.ColorBg)

	right := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight)
	right.SetBackgroundColor(theme.ColorBg)

	view := tview.NewFlex().SetDirection(tview.FlexColumn)
	view.SetBackgroundColor(theme.ColorBg)
	view.AddItem(left, 0, 1, false)
	view.AddItem(right, 0, 1, false)

	return &TitleBar{View: view, left: left, right: right}
}

func (t *TitleBar) Update(state *models.State) {
	scan := fmt.Sprintf("[#7d8590]Scanned[-] [#8b949e][[-][#f0883e]S[-][#8b949e]]:[-] [#8b949e]%d keys[-]",
		state.ScannedKeys,
	)
	monitor := fmt.Sprintf("[#7d8590]Monitor[-] [#8b949e][[-][#f0883e]M[-][#8b949e]]:[-] [#8b949e]%s[-]",
		utils.FormatDuration(int64(state.TotalMonitorDuration.Seconds())),
	)
	t.left.SetText(fmt.Sprintf(" [#58a6ff::b]RedScout[-::-]  %s  %s", scan, monitor))

	helpHint := "[#8b949e][[-][#f0883e]?[-][#8b949e]][-] [#8b949e]help[-]"
	status := fmt.Sprintf("[#3fb950]●[-] [#8b949e::d]%s[-::-]", state.Status)
	t.right.SetText(fmt.Sprintf("%s   %s ", helpHint, status))
}
