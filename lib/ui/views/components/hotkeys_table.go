package components

import (
	"fmt"
	"redscout/models"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const HotKeysShortcutsText = "[yellow]S[-] +SCAN  |  [yellow]M[-] +MONITOR  |  [yellow]Q[-] Quit"

func NewHotKeyTable() *tview.Table {
	table := tview.NewTable().SetFixed(1, 0)
	table.SetTitle(" Hot Keys (Top N by Ops) ").SetTitleAlign(tview.AlignLeft)
	table.SetSelectable(true, false)
	table.SetBorders(false)
	return table
}

func UpdateHotKeyTable(table *tview.Table, hotKeys models.HotKeyList) {
	headers := []string{"Key", "Ops"}
	colors := []tcell.Color{
		tcell.ColorWhite,
		tcell.ColorAqua,
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

	for i, row := range hotKeys {
		values := []string{
			row.Key.String(),
			fmt.Sprintf("%8.1f/s", row.Ops),
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
