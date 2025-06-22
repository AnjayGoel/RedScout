package components

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"redscout/lib/utils"
	"redscout/models"
)

const SpecialKeysShortcutsText = "[yellow]B[-] Toggle Table  |  [yellow]T[-] Toggle View  |  [yellow]S[-] +SCAN  |  [yellow]M[-] +MONITOR   |  [yellow]Q[-] Quit"

type SpecialKeysView struct {
	Flex        *tview.Flex
	bigKeyTable *tview.Table
	hotKeyTable *tview.Table
	app         *tview.Application
}

func NewSpecialKeysView(app *tview.Application) *SpecialKeysView {
	bigKeyTable := BuildBigKeyTable()
	hotKeyTable := BuildHotKeyTable()

	// Add left padding to tables
	bigKeyTable.SetOffset(0, 2)
	hotKeyTable.SetOffset(0, 2)

	flex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(bigKeyTable, 0, 1, true).
		AddItem(hotKeyTable, 0, 1, false)

	sv := &SpecialKeysView{
		Flex:        flex,
		bigKeyTable: bigKeyTable,
		hotKeyTable: hotKeyTable,
		app:         app,
	}

	// Track which table is focused
	focusedTable := 0 // 0 = bigKeyTable, 1 = hotKeyTable
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'B' || event.Rune() == 'b' {
			if focusedTable == 0 {
				sv.hotKeyTable.Select(1, 0)
				sv.app.SetFocus(sv.hotKeyTable)
				focusedTable = 1
			} else {
				sv.bigKeyTable.Select(1, 0)
				sv.app.SetFocus(sv.bigKeyTable)
				focusedTable = 0
			}
			return nil
		}
		return event
	})

	return sv
}

func (s *SpecialKeysView) Update(state *models.State) {
	PopulateBigKeyTable(s.bigKeyTable, state.BigKeys)
	PopulateHotKeyTable(s.hotKeyTable, state.HotKeys)
}

func (s *SpecialKeysView) Focus() {
	s.bigKeyTable.Select(1, 0)
	s.app.SetFocus(s.bigKeyTable)
}

func BuildBigKeyTable() *tview.Table {
	table := tview.NewTable().SetFixed(1, 0)
	table.SetTitle(" Big Keys (Top N by Memory) ").SetTitleAlign(tview.AlignLeft)
	table.SetSelectable(true, false)
	table.SetBorders(false)
	return table
}

func PopulateBigKeyTable(table *tview.Table, bigKeys models.BigKeyList) {
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

func BuildHotKeyTable() *tview.Table {
	table := tview.NewTable().SetFixed(1, 0)
	table.SetTitle(" Hot Keys (Top N by Ops) ").SetTitleAlign(tview.AlignLeft)
	table.SetSelectable(true, false)
	table.SetBorders(false)
	return table
}

func PopulateHotKeyTable(table *tview.Table, hotKeys models.HotKeyList) {
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
