package components

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"redmon/lib/utils"
	"redmon/models"
	"strings"
)

const StatsHeader = "[yellow]1[-] Keys  [yellow]2[-] Memory  [yellow]3[-] Avg TTL  [yellow]4[-] % TTL  [yellow]5[-] GET  [yellow]6[-] SET  [yellow]7[-] DEL  [yellow]8[-] OPS  |  [yellow]Enter/→[-] Drill Down  [yellow]Backspace/←[-] Level Up  |  [yellow]S[-] +SCAN  |  [yellow]M[-] +MONITOR |  [yellow]T[-] Toggle View  |  [yellow]Q[-] Quit"

func BuildStatsTable() *tview.Table {
	table := tview.NewTable().SetFixed(1, 0)
	table.SetTitle(" Namespace Stats (Press 1-8 to sort) ").SetTitleAlign(tview.AlignLeft)
	table.SetSelectable(true, false)
	table.SetBorders(false)
	return table
}

func PopulateStatsTable(table *tview.Table, stats models.NamespaceMetricList) {
	headers := []string{"Namespace", "~Keys", "~Memory", "Avg TTL", "% TTL", "GET/s", "SET/s", "DEL/s", "Total Ops/s", "Types"}
	colors := []tcell.Color{
		tcell.ColorWhite,
		tcell.ColorYellow,
		tcell.ColorAqua,
		tcell.ColorLightGreen,
		tcell.ColorLightCyan,
		tcell.ColorBlue,
		tcell.ColorGreen,
		tcell.ColorRed,
		tcell.ColorPurple,
		tcell.ColorGray,
	}

	// Calculate max width for each column
	colWidths := make([]int, len(headers))
	for i := range headers {
		// Consider the header width (12 is the format width we use)
		colWidths[i] = 12
	}

	// Add header row
	table.Clear()
	for i, h := range headers {
		align := tview.AlignLeft
		if i != 0 && i != (len(headers)-1) {
			align = tview.AlignRight
		}
		cell := tview.NewTableCell(fmt.Sprintf("[white::b]%s", h)).
			SetTextColor(tcell.ColorWhite).
			SetAttributes(tcell.AttrBold).
			SetBackgroundColor(tcell.ColorTeal).
			SetSelectable(false).
			SetAlign(align)
		table.SetCell(0, i, cell)
	}

	// Add data rows and update max widths
	for i, row := range stats {
		values := []string{
			fmt.Sprintf("%-20s", row.Namespace),
			fmt.Sprintf("%12s", utils.FormatNumber(float64(row.EstKeys))),
			fmt.Sprintf("%12s", utils.FormatBytes(row.EstMemory)),
			fmt.Sprintf("%12s", utils.FormatDuration(row.AvgTTL)),
			fmt.Sprintf("%11.1f%%", row.TTLPercent*100),
			fmt.Sprintf("%8.1f/s", row.Ops[models.GetOp]),
			fmt.Sprintf("%8.1f/s", row.Ops[models.SetOp]),
			fmt.Sprintf("%8.1f/s", row.Ops[models.DelOp]),
			fmt.Sprintf("%8.1f/s", row.Ops[models.TotalOp]),
			fmt.Sprintf("%-12s", strings.Join(row.Types[:], ",")),
		}

		// Update max widths
		for j, val := range values {
			contentWidth := len(strings.TrimSpace(val))
			if contentWidth > colWidths[j] {
				colWidths[j] = contentWidth
			}
		}

		for j, val := range values {
			align := tview.AlignLeft
			if j != 0 && j != (len(headers)-1) {
				align = tview.AlignRight
			}
			cell := tview.NewTableCell(fmt.Sprintf("[%s]%s", colors[j], val)).
				SetAlign(align).
				SetExpansion(0).
				SetBackgroundColor(tcell.ColorBlack)

			table.SetCell(i+1, j, cell)
		}
	}

	// Set column widths
	totalWidth := 0
	for i, width := range colWidths {
		width += 6
		table.SetCell(0, i, table.GetCell(0, i).SetExpansion(0))
		table.SetCell(0, i, table.GetCell(0, i).SetMaxWidth(width))
		totalWidth += width
	}

	// Set statsTable width to total width of columns
	table.SetFixed(1, 0)
	table.ScrollToBeginning()
}
