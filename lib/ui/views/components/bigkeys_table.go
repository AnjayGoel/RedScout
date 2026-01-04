package components

import (
	"fmt"
	"redscout/lib/utils"
	"redscout/models"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const BigKeysShortcutsText = "[yellow]S[-] +SCAN  |  [yellow]M[-] +MONITOR  |  [yellow]Q[-] Quit"

func NewBigKeyTable() *tview.Table {
	table := tview.NewTable().SetFixed(1, 0)
	table.SetTitle(" Big Keys (Top N by Memory) ").SetTitleAlign(tview.AlignLeft)
	table.SetSelectable(true, false)
	table.SetBorders(false)
	return table
}

func UpdateBigKeyTable(table *tview.Table, bigKeys models.BigKeyList) {
	headers := []string{"Key", "Size"}
	colors := []tcell.Color{
		tcell.ColorWhite,
		tcell.ColorYellow,
	}

	table.Clear()
	for i, h := range headers {
		cell := tview.NewTableCell(fmt.Sprintf("[white::b]%s", h)).
			SetTextColor(tcell.ColorWhite).
			SetAttributes(tcell.AttrBold).
			SetBackgroundColor(tcell.ColorAqua).
			SetSelectable(false).
			SetAlign(tview.AlignLeft)
		table.SetCell(0, i, cell)
	}

	for i, row := range bigKeys {
		values := []string{
			row.Key.String(),
			fmt.Sprintf("%12s", utils.FormatBytes(row.Size)),
		}
		for j, val := range values {
			cell := tview.NewTableCell(fmt.Sprintf("[%s]%s", colors[j], val)).
				SetAlign(tview.AlignLeft).
				SetExpansion(0).
				SetBackgroundColor(tcell.ColorBlack)
			table.SetCell(i+1, j, cell)
		}
	}
	table.ScrollToBeginning()
}
