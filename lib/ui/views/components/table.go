package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"redscout/lib/ui/theme"
)

type column struct {
	header string
	align  int
	color  tcell.Color
}

func newTable() *tview.Table {
	table := tview.NewTable().SetFixed(1, 0)
	table.SetSelectable(true, false)
	table.SetBorderPadding(0, 0, 1, 0)
	table.SetBackgroundColor(theme.ColorBg)
	table.SetSelectedStyle(tcell.StyleDefault.
		Background(theme.ColorHighlight).
		Foreground(theme.ColorText).
		Attributes(tcell.AttrBold))
	return table
}

func renderTable(table *tview.Table, columns []column, rows [][]string) {
	table.Clear()
	for j, c := range columns {
		table.SetCell(0, j, tview.NewTableCell(c.header).
			SetBackgroundColor(theme.ColorBg).
			SetSelectable(false).
			SetAlign(c.align))
	}

	for i, values := range rows {
		rowBg := theme.ColorBg
		if i%2 == 1 {
			rowBg = theme.ColorSurface
		}
		for j, val := range values {
			table.SetCell(i+1, j, tview.NewTableCell(val).
				SetTextColor(columns[j].color).
				SetAlign(columns[j].align).
				SetBackgroundColor(rowBg))
		}
	}
}
