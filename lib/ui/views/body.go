package views

import (
	"github.com/rivo/tview"
	"redscout/lib/ui/views/components"
	"redscout/models"
)

type Tab string

const (
	TabNamespace Tab = "namespace"
	TabSlowLog   Tab = "slowlog"
	TabBigKeys   Tab = "bigkeys"
	TabHotKeys   Tab = "hotkeys"
)

type BodyView struct {
	Shortcuts   *tview.TextView
	ContentFlex *tview.Flex
	namespace   *components.Namespace
	slowLog     *components.SlowLogTable
	activeView  Tab
	TabBar      *tview.TextView
	app         *tview.Application
	bigKeyTable *tview.Table
	hotKeyTable *tview.Table
}

func NewBodyView(app *tview.Application) *BodyView {
	view := &BodyView{
		app:         app,
		Shortcuts:   newShortcuts(),
		ContentFlex: newContentFlex(),
		namespace:   components.NewNamespace(),
		slowLog:     components.NewSlowLogTable(),
		activeView:  TabNamespace,
		TabBar:      newTabBar(),
		bigKeyTable: components.NewBigKeyTable(),
		hotKeyTable: components.NewHotKeyTable(),
	}
	view.SetActiveView(TabNamespace)
	return view
}

var namespaceSortKeyMap = map[rune]string{
	'1': "Keys",
	'2': "Memory",
	'3': "TTL",
	'4': "% TTL",
	'5': "Get",
	'6': "Set",
	'7': "Del",
	'8': "Total Ops",
}

var slowLogSortKeyMap = map[rune]string{
	'1': "ID",
	'2': "Timestamp",
	'3': "Duration",
	'4': "Command",
}

func newShortcuts() *tview.TextView {
	return tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)
}

func newContentFlex() *tview.Flex {
	return tview.NewFlex().SetDirection(tview.FlexColumn)
}

func newTabBar() *tview.TextView {
	tabBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	return tabBar
}

func (b *BodyView) SetActiveView(view Tab) {
	b.activeView = view

	switch view {
	case TabNamespace:
		b.ContentFlex.Clear().AddItem(b.namespace.Flex, 0, 2, true)
		b.Shortcuts.SetText(components.StatsHeader)
		b.TabBar.SetText(
			`[::b][white:teal][[yellow]N[-]]amespace [white::-]` +
				`[white:black][-:-]` +
				`[white] Slow [[yellow]L[-]]og [-]` +
				`[white:black][-:-]` +
				`[white] [[yellow]B[-]]ig Keys [-]` +
				`[white:black][-:-]` +
				`[white] [[yellow]H[-]]ot Keys [-]`,
		)
		b.namespace.Table.Select(1, 0)
		b.app.SetFocus(b.namespace.Table)
	case TabSlowLog:
		b.ContentFlex.Clear().AddItem(b.slowLog.Table, 0, 2, true)
		b.Shortcuts.SetText(components.SlowLogHeader)
		b.TabBar.SetText(
			`[white][[yellow]N[-]]amespace [-]` +
				`[white:black][-:-]` +
				`[::b][white:teal] Slow [[yellow]L[-]]og [white::-]` +
				`[white:black][-:-]` +
				`[white] [[yellow]B[-]]ig Keys [-]` +
				`[white:black][-:-]` +
				`[white] [[yellow]H[-]]ot Keys [-]`,
		)
		b.slowLog.Table.Select(1, 0)
		b.app.SetFocus(b.slowLog.Table)
	case TabBigKeys:
		b.ContentFlex.Clear().AddItem(b.bigKeyTable, 0, 2, true)
		b.Shortcuts.SetText(components.BigKeysShortcutsText)
		b.TabBar.SetText(
			`[white][[yellow]N[-]]amespace [-]` +
				`[white:black][-:-]` +
				`[white] Slow [[yellow]L[-]]og [-]` +
				`[white:black][-:-]` +
				`[::b][white:teal] [[yellow]B[-]]ig Keys [white::-]` +
				`[white:black][-:-]` +
				`[white] [[yellow]H[-]]ot Keys [-]`,
		)
		b.bigKeyTable.Select(1, 0)
		b.app.SetFocus(b.bigKeyTable)
	case TabHotKeys:
		b.ContentFlex.Clear().AddItem(b.hotKeyTable, 0, 2, true)
		b.Shortcuts.SetText(components.HotKeysShortcutsText)
		b.TabBar.SetText(
			`[white][[yellow]N[-]]amespace [-]` +
				`[white:black][-:-]` +
				`[white] Slow [[yellow]L[-]]og [-]` +
				`[white:black][-:-]` +
				`[white] [[yellow]B[-]]ig Keys [-]` +
				`[white:black][-:-]` +
				`[::b][white:teal] [[yellow]H[-]]ot Keys [white::-]`,
		)
		b.hotKeyTable.Select(1, 0)
		b.app.SetFocus(b.hotKeyTable)
	}
}

func (b *BodyView) ToggleView() {
	switch b.activeView {
	case TabNamespace:
		b.SetActiveView(TabSlowLog)
	case TabSlowLog:
		b.SetActiveView(TabBigKeys)
	case TabBigKeys:
		b.SetActiveView(TabHotKeys)
	default:
		b.SetActiveView(TabNamespace)
	}
}

func (b *BodyView) Update(data *models.State) {
	b.slowLog.Update(data.SlowLogs)
	b.namespace.Update(data.CurrentPrefix, data.NamespaceStats)
	components.UpdateBigKeyTable(b.bigKeyTable, data.BigKeys)
	components.UpdateHotKeyTable(b.hotKeyTable, data.HotKeys)
}

func (b *BodyView) HandleInput(inp rune, state *models.State) {
	if inp == 'T' || inp == 't' {
		b.ToggleView()
		return
	}
	if inp == 'B' || inp == 'b' {
		b.SetActiveView(TabBigKeys)
		return
	}
	if inp == 'H' || inp == 'h' {
		b.SetActiveView(TabHotKeys)
		return
	}
	if inp == 'N' || inp == 'n' {
		b.SetActiveView(TabNamespace)
		return
	}
	if inp == 'L' || inp == 'l' {
		b.SetActiveView(TabSlowLog)
		return
	}
	if inp > '8' || inp < '1' {
		return
	}
	key := ""
	if b.activeView == TabNamespace {
		key = namespaceSortKeyMap[inp]
		if key == "" {
			return
		}
		state.NamespaceStats.Sort(key)
	} else if b.activeView == TabSlowLog {
		key = slowLogSortKeyMap[inp]
		if key == "" {
			return
		}
		state.SlowLogs.Sort(key)
	}
	b.Update(state)
}

func (b *BodyView) ActiveView() Tab {
	return b.activeView
}

func (b *BodyView) NamespaceTable() *tview.Table {
	return b.namespace.Table
}
