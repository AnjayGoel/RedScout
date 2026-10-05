package components

import (
	"fmt"
	"slices"
	"strings"

	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"

	"github.com/rivo/tview"
)

type Namespace struct {
	Title      *tview.TextView
	Table      *tview.Table
	Flex       *tview.Flex
	lastPrefix models.Key
}

var namespaceColumns = []column{
	{"[#8b949e]Namespace[-]", tview.AlignLeft, theme.ColorText},
	{"[#8b949e]~Keys [#7d8590][[-][#f0883e]1[-][#7d8590]][-]", tview.AlignRight, theme.ColorOrange},
	{"[#8b949e]~Memory [#7d8590][[-][#f0883e]2[-][#7d8590]][-]", tview.AlignRight, theme.ColorBlue},
	{"[#8b949e]Avg TTL [#7d8590][[-][#f0883e]3[-][#7d8590]][-]", tview.AlignRight, theme.ColorSecondary},
	{"[#8b949e]% TTL [#7d8590][[-][#f0883e]4[-][#7d8590]][-]", tview.AlignRight, theme.ColorSecondary},
	{"[#8b949e]GET/s [#7d8590][[-][#f0883e]5[-][#7d8590]][-]", tview.AlignRight, theme.ColorSecondary},
	{"[#8b949e]SET/s [#7d8590][[-][#f0883e]6[-][#7d8590]][-]", tview.AlignRight, theme.ColorSecondary},
	{"[#8b949e]DEL/s [#7d8590][[-][#f0883e]7[-][#7d8590]][-]", tview.AlignRight, theme.ColorSecondary},
	{"[#8b949e]OPS/s [#7d8590][[-][#f0883e]8[-][#7d8590]][-]", tview.AlignRight, theme.ColorSecondary},
	{"[#8b949e]Types[-]", tview.AlignLeft, theme.ColorMuted},
}

func NewNamespace() *Namespace {
	ns := &Namespace{Table: newTable()}

	ns.Title = tview.NewTextView()
	ns.Title.SetDynamicColors(true)
	ns.Title.SetBackgroundColor(theme.ColorBg)
	ns.Title.SetText(" [#f0883e]/ root[-]")

	ns.Flex = tview.NewFlex()
	ns.Flex.SetDirection(tview.FlexRow)
	ns.Flex.SetBackgroundColor(theme.ColorBg)
	ns.Flex.AddItem(ns.Title, 2, -1, false)
	ns.Flex.AddItem(ns.Table, 0, 1, true)

	return ns
}

func (ns *Namespace) Update(prefix models.Key, stats models.NamespaceMetricList) {
	rows := make([][]string, len(stats))
	for i, row := range stats {
		rows[i] = []string{
			fmt.Sprintf("%-*s", utils.MaxKeyDisplayLen, utils.TruncateKey(row.Namespace)),
			fmt.Sprintf("%12s", utils.FormatNumber(float64(row.EstKeys))),
			fmt.Sprintf("%12s", utils.FormatBytes(row.EstMemory)),
			fmt.Sprintf("%12s", utils.FormatDuration(row.AvgTTL)),
			fmt.Sprintf("%11.1f%%", row.TTLPercent*100),
			fmt.Sprintf("%10.1f/s", row.Ops[models.GetOp]),
			fmt.Sprintf("%10.1f/s", row.Ops[models.SetOp]),
			fmt.Sprintf("%10.1f/s", row.Ops[models.DelOp]),
			fmt.Sprintf("%10.1f/s", row.Ops[models.TotalOp]),
			strings.Join(row.Types, ","),
		}
	}
	renderTable(ns.Table, namespaceColumns, rows)

	// Reset selection only when prefix changes (drill-down/back)
	if !slices.Equal(prefix, ns.lastPrefix) {
		ns.Table.Select(selectedRowAfterMove(ns.lastPrefix, prefix, stats), 0)
		ns.lastPrefix = prefix
	}

	separator := " › "
	hints := "  [#7d8590]([[-][#f0883e]→[-][#7d8590]][-] [#7d8590]expand  [[-][#f0883e]←[-][#7d8590]][-] [#7d8590]back)[-]"
	line := "\n[#21262d]" + strings.Repeat("─", 300) + "[-]"
	if len(prefix) == 0 {
		ns.Title.SetText(" [#f0883e]/ root[-]" + hints + line)
	} else {
		path := "/ root" + separator + strings.Join(prefix, separator)
		ns.Title.SetText(fmt.Sprintf(" [#f0883e]%s[-]%s%s", path, hints, line))
	}
}

func selectedRowAfterMove(from, to models.Key, stats models.NamespaceMetricList) int {
	if len(from) == len(to)+1 && slices.Equal(from[:len(to)], to) {
		for i, row := range stats {
			if row.Namespace == from[len(to)] {
				return i + 1
			}
		}
	}
	return 1
}
