package components

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"
)

var slowLogColumns = []column{
	{"[#8b949e]ID [#7d8590][[-][#f0883e]1[-][#7d8590]][-]", tview.AlignRight, theme.ColorText},
	{"[#8b949e]Timestamp [#7d8590][[-][#f0883e]2[-][#7d8590]][-]", tview.AlignLeft, theme.ColorSecondary},
	{"[#8b949e]Duration [#7d8590][[-][#f0883e]3[-][#7d8590]][-]", tview.AlignRight, theme.ColorOrange},
	{"[#8b949e]Command [#7d8590][[-][#f0883e]4[-][#7d8590]][-]", tview.AlignLeft, theme.ColorGreen},
	{"[#8b949e]Arguments[-]", tview.AlignLeft, theme.ColorMuted},
}

type SlowLogTable struct {
	Table *tview.Table
}

func NewSlowLogTable() *SlowLogTable {
	return &SlowLogTable{Table: newTable()}
}

func (sl *SlowLogTable) Update(slowLogs models.SlowLogList) {
	rows := make([][]string, len(slowLogs))
	for i, entry := range slowLogs {
		command := ""
		var args []string
		if len(entry.Args) > 0 {
			command = strings.ToUpper(entry.Args[0])
			args = entry.Args[1:]
		}
		rows[i] = []string{
			fmt.Sprintf("%d ", entry.ID),
			entry.Time.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%12d ms", entry.Duration.Milliseconds()),
			command,
			tview.Escape(utils.TruncateKey(strings.Join(args, " "))),
		}
	}
	renderTable(sl.Table, slowLogColumns, rows)
}
