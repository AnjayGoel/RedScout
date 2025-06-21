package views

import (
	"fmt"
	"github.com/rivo/tview"
	"redmon/lib/ui/views/components"
	"redmon/lib/utils"
	"redmon/models"
	"time"
)

type HeaderView struct {
	HeaderFlex  *tview.Flex
	logs        *tview.TextView
	system      *tview.TextView
	performance *tview.TextView
	resources   *tview.TextView
}

func NewHeaderView() *HeaderView {
	system := tview.NewTextView().SetDynamicColors(true)
	system.SetBorder(true).SetTitle("[teal]System Info[-]")

	stats := tview.NewTextView().SetDynamicColors(true)
	stats.SetBorder(true).SetTitle("[teal]Performance[-]")

	memory := tview.NewTextView().SetDynamicColors(true)
	memory.SetBorder(true).SetTitle("[teal]Resources[-]")

	headerFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	logs := tview.NewTextView().SetDynamicColors(true)
	logs.SetBorder(true).SetTitle("[teal]Scan State[-]")

	headerFlex.AddItem(system, 0, 1, false)
	headerFlex.AddItem(stats, 0, 1, false)
	headerFlex.AddItem(memory, 0, 1, false)
	headerFlex.AddItem(logs, 0, 1, false)

	return &HeaderView{
		system:      system,
		performance: stats,
		resources:   memory,
		HeaderFlex:  headerFlex,
		logs:        logs,
	}
}

func (header *HeaderView) Update(state *models.State) {
	header.updateHeaderSystemView(state.RedisInfo)
	header.updateHeaderPerformanceView(state.RedisInfo)
	header.updateHeaderResourcesView(state.RedisInfo)
	header.updateLogs(state)
}

func (header *HeaderView) updateHeaderSystemView(info *models.RedisInfo) {
	uptime := time.Duration(info.Server.Uptime) * time.Second
	text := fmt.Sprintf(" [teal]Redis: [-][white]v%s[-]\n [teal]OS:[-][white] %s[-]\n [teal]Uptime:[-][white] %s[-]\n [teal]Clients:[-][white] %d[-]",
		info.Server.RedisVersion,
		info.Server.OS,
		utils.FormatDuration(int64(uptime.Seconds())),
		info.Clients.ConnectedClients,
	)
	header.system.SetText(text)
}

func (header *HeaderView) updateHeaderPerformanceView(info *models.RedisInfo) {
	totalKeys := info.Keyspace["db0"].Keys
	avgTTL := info.Keyspace["db0"].AvgTTL

	text := fmt.Sprintf(" [teal]Total Keys:[-][white] %s[-]\n [teal]Ops:[-][white] %s[-]\n [teal]Hit Rate:[-][white] %.1f%%[-]\n [teal]Avg TTL:[-][white] %s[-]",
		utils.FormatNumber(float64(totalKeys)),
		utils.FormatOpsPerSec(float64(info.Stats.OpsPerSec)),
		info.Computed.HitRate*100,
		utils.FormatDuration(avgTTL),
	)
	header.performance.SetText(text)
}

func (header *HeaderView) updateHeaderResourcesView(info *models.RedisInfo) {
	// Calculate resources usage percentage
	memPercent := 0.0
	if info.Memory.MaxMemory > 0 {
		memPercent = float64(info.Memory.UsedMemory) / float64(info.Memory.MaxMemory) * 100
	}

	// Create progress bars
	cpuBar := components.CreateProgressBar(info.Computed.CPUUsage*100, 100, 20)
	memBar := components.CreateProgressBar(memPercent, 100, 20)

	text := fmt.Sprintf(" [teal]CPU:[-][white] %s[-]\n [teal]Memory:[-][white] %s[-]\n [teal]Max Mem:[-][white] %s[-]\n [teal]Eviction Policy:[-][white] %s[-]",
		cpuBar,
		memBar,
		info.Memory.MaxMemoryHuman,
		info.Memory.MemoryPolicy,
	)
	header.resources.SetText(text)
}

func (header *HeaderView) updateLogs(state *models.State) {
	text := fmt.Sprintf(
		" [teal]Keys Scanned:[-] %d\n [teal]Monitored Duration:[-] %s\n [teal]Scan Cursor:[-] %d\n [teal]Logs:[-] %s\n",
		state.ScannedKeys,
		utils.FormatDuration(int64(state.TotalMonitorDuration.Seconds())),
		state.Cursor,
		state.Status,
	)
	header.logs.SetText(text)
}
