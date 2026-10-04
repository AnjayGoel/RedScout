package views

import (
	"fmt"

	"github.com/rivo/tview"
	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"
)

type HeaderView struct {
	HeaderFlex  *tview.Flex
	system      *tview.TextView
	performance *tview.TextView
	resources   *tview.TextView
}

func NewHeaderView() *HeaderView {
	system := tview.NewTextView().SetDynamicColors(true)
	system.SetBorder(true).
		SetBorderColor(theme.ColorBorder).
		SetBackgroundColor(theme.ColorBg)

	perf := tview.NewTextView().SetDynamicColors(true)
	perf.SetBorder(true).
		SetBorderColor(theme.ColorBorder).
		SetBackgroundColor(theme.ColorBg)

	mem := tview.NewTextView().SetDynamicColors(true)
	mem.SetBorder(true).
		SetBorderColor(theme.ColorBorder).
		SetBackgroundColor(theme.ColorBg)

	headerFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	headerFlex.SetBackgroundColor(theme.ColorBg)
	headerFlex.AddItem(system, 0, 1, false)
	headerFlex.AddItem(perf, 0, 1, false)
	headerFlex.AddItem(mem, 0, 1, false)

	return &HeaderView{
		system:      system,
		performance: perf,
		resources:   mem,
		HeaderFlex:  headerFlex,
	}
}

func (header *HeaderView) Update(state *models.State) {
	header.updateSystem(state.RedisInfo)
	header.updatePerformance(state.RedisInfo, state.Keyspace())
	header.updateMemory(state.RedisInfo)
}

func (header *HeaderView) updateSystem(info *models.RedisInfo) {
	os := info.Server.OS
	if len(os) > 20 {
		os = os[:20]
	}
	text := fmt.Sprintf(" [#7d8590]SYSTEM[-]\n [#58a6ff]Redis v%s[-] [#7d8590]· %s[-]\n [#7d8590]Uptime[-] [#8b949e]%s[-] [#7d8590]· Clients[-] [#8b949e]%d[-]",
		info.Server.RedisVersion,
		os,
		utils.FormatUptime(info.Server.Uptime),
		info.Clients.ConnectedClients,
	)
	header.system.SetText(text)
}

func (header *HeaderView) updatePerformance(info *models.RedisInfo, keyspace models.KeyspaceInfo) {
	totalKeys := keyspace.Keys
	avgTTL := keyspace.AvgTTL

	text := fmt.Sprintf(" [#7d8590]PERFORMANCE[-]\n [#7d8590]Keys[-] [#f0883e]%s[-] [#7d8590]· Ops[-] [#f0883e]%s[-]\n [#7d8590]Hit[-] [#3fb950]%.1f%%[-] [#7d8590]· Avg TTL[-] [#8b949e]%s[-]",
		utils.FormatNumber(float64(totalKeys)),
		utils.FormatOpsPerSec(float64(info.Stats.OpsPerSec)),
		info.Computed.HitRate*100,
		utils.FormatDuration(avgTTL),
	)
	header.performance.SetText(text)
}

func (header *HeaderView) updateMemory(info *models.RedisInfo) {
	// Memory: used / max (percent) or used (no limit)
	var memLine string
	if info.Memory.MaxMemory > 0 {
		memPercent := float64(info.Memory.UsedMemory) / float64(info.Memory.MaxMemory) * 100
		memLine = fmt.Sprintf("[#7d8590]Mem[-] [#f0883e::b]%s[-::-] [#7d8590]/[-] [#8b949e]%s[-] [#7d8590]([#f0883e]%.1f%%[-][#7d8590])[-]",
			info.Memory.UsedMemoryHuman,
			info.Memory.MaxMemoryHuman,
			memPercent,
		)
	} else {
		memLine = fmt.Sprintf("[#7d8590]Mem[-] [#f0883e::b]%s[-::-] [#7d8590]· no limit[-]", info.Memory.UsedMemoryHuman)
	}

	// CPU
	var cpuPart string
	if info.CPU.SystemTime == 0 && info.CPU.UserTime == 0 {
		cpuPart = "[#7d8590]CPU[-] [#8b949e]N/A[-]"
	} else {
		cpuPart = fmt.Sprintf("[#7d8590]CPU[-] [#3fb950]%.1f%%[-]", info.Computed.CPUUsage*100)
	}

	text := fmt.Sprintf(" [#7d8590]RESOURCES[-]\n %s [#7d8590]·[-] [#8b949e]%s[-]\n %s",
		memLine,
		info.Memory.MemoryPolicy,
		cpuPart,
	)
	header.resources.SetText(text)
}
