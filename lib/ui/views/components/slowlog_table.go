package components

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"redscout/models"
	"strings"
)

const SlowLogHeader = "[yellow]1[-] ID  [yellow]2[-] Timestamp  [yellow]3[-] Duration  [yellow]4[-] Command  [yellow] |  [yellow]S[-] +SCAN  |  [yellow]M[-] +MONITOR |  [yellow]T[-] Toggle View  |  [yellow]Q[-] Quit"

func BuildSlowLogTable() *tview.Table {
	table := tview.NewTable().SetFixed(1, 0)
	table.SetTitle(" Slow Log (Press 1-5 to sort) ").SetTitleAlign(tview.AlignLeft)
	table.SetSelectable(true, false)
	table.SetBorders(false)
	return table
}

func PopulateSlowLogTable(table *tview.Table, slowLogs models.SlowLogList) {
	if slowLogs == nil || len(slowLogs) == 0 {
		return
	}

	headers := []string{"ID", "Timestamp", "Duration", "Command", "Arguments"}
	colors := []tcell.Color{
		tcell.ColorWhite,
		tcell.ColorYellow,
		tcell.ColorTeal,
		tcell.ColorLightGreen,
		tcell.ColorBlue,
	}

	// Set minimum column widths
	colWidths := []int{8, 20, 12, 12, 30}

	// Add header row
	table.Clear()
	for i, h := range headers {
		align := tview.AlignLeft
		if i != 0 && i != 4 {
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

	// Add data rows
	for i, log := range slowLogs {
		// Split command and arguments
		command := ""
		var args []string
		if len(log.Args) > 0 {
			command = strings.ToUpper(log.Args[0])
			args = log.Args[1:]
		}

		values := []string{
			fmt.Sprintf("%d ", log.ID),
			log.Time.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%12d ms", log.Duration.Milliseconds()),
			command,
			strings.Join(args, " "),
		}

		// Update max widths based on content
		for j, val := range values {
			contentWidth := len(val)
			if contentWidth > colWidths[j] {
				colWidths[j] = contentWidth
			}
		}

		for j, val := range values {
			align := tview.AlignLeft
			if j != 0 && j != 4 {
				align = tview.AlignRight
			}
			cell := tview.NewTableCell(fmt.Sprintf("[%s]%s", colors[j], val)).
				SetAlign(align).
				SetExpansion(0).
				SetBackgroundColor(tcell.ColorBlack)

			table.SetCell(i+1, j, cell)
		}
	}

	// Set column widths with padding
	totalWidth := 0
	for i, width := range colWidths {
		width += 2 // Add padding
		table.SetCell(0, i, table.GetCell(0, i).SetExpansion(0))
		table.SetCell(0, i, table.GetCell(0, i).SetMaxWidth(width))
		totalWidth += width
	}

	// Set table width to total width of columns
	table.SetFixed(1, 0)
}
