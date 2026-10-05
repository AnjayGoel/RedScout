package components

import (
	"fmt"

	"github.com/rivo/tview"

	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"
)

var hotKeyColumns = []column{
	{"[#8b949e]Key[-]", tview.AlignLeft, theme.ColorText},
	{"[#8b949e]Ops/s[-]", tview.AlignRight, theme.ColorBlue},
	{"[#8b949e]Command[-]", tview.AlignLeft, theme.ColorSecondary},
}

func NewHotKeyTable() *tview.Table {
	return newTable()
}

func UpdateHotKeyTable(table *tview.Table, hotKeys models.HotKeyList) {
	rows := make([][]string, len(hotKeys))
	for i, row := range hotKeys {
		rows[i] = []string{
			fmt.Sprintf("%-*s", utils.MaxKeyDisplayLen, utils.TruncateKey(row.Key)),
			fmt.Sprintf("%10.1f/s", row.Ops),
			fmt.Sprintf("%-8s", row.Command),
		}
	}
	renderTable(table, hotKeyColumns, rows)
}
