package components

import (
	"fmt"

	"github.com/rivo/tview"

	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"
)

var bigKeyColumns = []column{
	{"[#8b949e]Key[-]", tview.AlignLeft, theme.ColorText},
	{"[#8b949e]Size[-]", tview.AlignRight, theme.ColorOrange},
	{"[#8b949e]Type[-]", tview.AlignRight, theme.ColorSecondary},
	{"[#8b949e]TTL[-]", tview.AlignRight, theme.ColorGreen},
}

func NewBigKeyTable() *tview.Table {
	return newTable()
}

func UpdateBigKeyTable(table *tview.Table, bigKeys models.BigKeyList) {
	rows := make([][]string, len(bigKeys))
	for i, row := range bigKeys {
		ttl := "-"
		if row.TTL > 0 {
			ttl = utils.FormatDuration(row.TTL)
		}
		rows[i] = []string{
			fmt.Sprintf("%-*s", utils.MaxKeyDisplayLen, utils.TruncateKey(row.Key)),
			fmt.Sprintf("%12s", utils.FormatBytes(row.Size)),
			fmt.Sprintf("%8s", row.Type),
			fmt.Sprintf("%12s", ttl),
		}
	}
	renderTable(table, bigKeyColumns, rows)
}
