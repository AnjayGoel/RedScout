package components

import (
	"fmt"
	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func NewBigKeyTable() *tview.Table {
	table := tview.NewTable().SetFixed(1, 0)
	table.SetSelectable(true, false)
	table.SetBorders(false)
	table.SetBorderPadding(0, 0, 1, 0)
	table.SetBackgroundColor(theme.ColorBg)
	table.SetSelectedStyle(tcell.StyleDefault.
		Background(theme.ColorHighlight).
		Foreground(theme.ColorText).
		Attributes(tcell.AttrBold))
	return table
}

func UpdateBigKeyTable(table *tview.Table, bigKeys models.BigKeyList) {
	headers := []string{
		"[#8b949e]Key[-]",
		"[#8b949e]Size [#7d8590][[-][#f0883e]1[-][#7d8590]][-]",
		"[#8b949e]Type [#7d8590][[-][#f0883e]2[-][#7d8590]][-]",
		"[#8b949e]TTL [#7d8590][[-][#f0883e]3[-][#7d8590]][-]",
	}

	table.Clear()
	for i, h := range headers {
		align := tview.AlignLeft
		if i >= 1 {
			align = tview.AlignRight
		}
		cell := tview.NewTableCell(h).
			SetBackgroundColor(theme.ColorBg).
			SetSelectable(false).
			SetAlign(align)
		table.SetCell(0, i, cell)
	}

	for i, row := range bigKeys {
		ttlStr := "-"
		if row.TTL > 0 {
			ttlStr = utils.FormatDuration(row.TTL)
		}

		values := []string{
			fmt.Sprintf("%-*s", utils.MaxKeyDisplayLen, utils.TruncateKey(row.Key.String())),
			fmt.Sprintf("%12s", utils.FormatBytes(row.Size)),
			fmt.Sprintf("%8s", row.Type),
			fmt.Sprintf("%12s", ttlStr),
		}

		colors := []tcell.Color{
			theme.ColorText,
			theme.ColorOrange,
			theme.ColorSecondary,
			theme.ColorGreen,
		}

		rowBg := theme.ColorBg
		if i%2 == 1 {
			rowBg = theme.ColorSurface
		}

		for j, val := range values {
			align := tview.AlignLeft
			if j >= 1 {
				align = tview.AlignRight
			}
			cell := tview.NewTableCell(val).
				SetTextColor(colors[j]).
				SetAlign(align).
				SetExpansion(0).
				SetBackgroundColor(rowBg)
			table.SetCell(i+1, j, cell)
		}
	}
}
