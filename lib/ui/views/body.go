package views

import (
	"fmt"

	"github.com/rivo/tview"

	"redscout/lib/ui/theme"
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

var tabOrder = []Tab{TabNamespace, TabSlowLog, TabBigKeys, TabHotKeys}

var tabShortcuts = map[rune]Tab{
	'n': TabNamespace,
	'l': TabSlowLog,
	'b': TabBigKeys,
	'h': TabHotKeys,
}

var namespaceSortColumns = map[rune]string{
	'1': "Keys",
	'2': "Memory",
	'3': "TTL",
	'4': "% TTL",
	'5': "Get",
	'6': "Set",
	'7': "Del",
	'8': "Total Ops",
}

var slowLogSortColumns = map[rune]string{
	'1': "ID",
	'2': "Timestamp",
	'3': "Duration",
	'4': "Command",
}

type BodyView struct {
	ContentFlex  *tview.Flex
	TabBar       *tview.TextView
	app          *tview.Application
	namespace    *components.Namespace
	slowLog      *components.SlowLogTable
	bigKeyTable  *tview.Table
	hotKeyTable  *tview.Table
	keyValueView *components.KeyValueView
	activeView   Tab
	showingValue bool
}

func NewBodyView(app *tview.Application) *BodyView {
	contentFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	contentFlex.SetBackgroundColor(theme.ColorBg)

	tabBar := tview.NewTextView().SetDynamicColors(true)
	tabBar.SetBackgroundColor(theme.ColorBg)

	view := &BodyView{
		ContentFlex:  contentFlex,
		TabBar:       tabBar,
		app:          app,
		namespace:    components.NewNamespace(),
		slowLog:      components.NewSlowLogTable(),
		bigKeyTable:  components.NewBigKeyTable(),
		hotKeyTable:  components.NewHotKeyTable(),
		keyValueView: components.NewKeyValueView(),
	}
	view.SetActiveView(TabNamespace)
	return view
}

func tabLabel(name string, key string, active bool) string {
	shortcut := fmt.Sprintf("[#7d8590][[-][#f0883e]%s[-][#7d8590]][-]", key)
	if active {
		// Active tab: bright blue + bold so the current view stands out.
		return fmt.Sprintf("[#58a6ff::b]%s[-::-] %s", name, shortcut)
	}
	return fmt.Sprintf("[#8b949e]%s[-] %s", name, shortcut)
}

func (b *BodyView) updateTabBar() {
	// Vertical dividers between tabs (not on the outer edges) so the row
	// reads as a bar of selectable tabs.
	div := "[#7d8590]│[-]"
	b.TabBar.SetText(fmt.Sprintf(" %s %s %s %s %s %s %s",
		tabLabel("Namespaces", "N", b.activeView == TabNamespace),
		div, tabLabel("Slow Log", "L", b.activeView == TabSlowLog),
		div, tabLabel("Big Keys", "B", b.activeView == TabBigKeys),
		div, tabLabel("Hot Keys", "H", b.activeView == TabHotKeys)))
}

func (b *BodyView) SetActiveView(view Tab) {
	b.showView(view, true)
}

func (b *BodyView) HideKeyValue() {
	b.showView(b.activeView, false)
}

func (b *BodyView) showView(view Tab, resetSelection bool) {
	b.activeView = view
	b.showingValue = false
	b.updateTabBar()

	table := b.activeTable()
	var content tview.Primitive = table
	if view == TabNamespace {
		content = b.namespace.Flex
	}

	b.ContentFlex.Clear().AddItem(content, 0, 1, true)
	if resetSelection {
		table.Select(1, 0)
	}
	b.app.SetFocus(table)
}

func (b *BodyView) activeTable() *tview.Table {
	switch b.activeView {
	case TabSlowLog:
		return b.slowLog.Table
	case TabBigKeys:
		return b.bigKeyTable
	case TabHotKeys:
		return b.hotKeyTable
	default:
		return b.namespace.Table
	}
}

func (b *BodyView) nextTab() Tab {
	for i, tab := range tabOrder {
		if tab == b.activeView {
			return tabOrder[(i+1)%len(tabOrder)]
		}
	}
	return TabNamespace
}

func (b *BodyView) Update(state *models.State) {
	b.slowLog.Update(state.SlowLogs)
	b.namespace.Update(state.CurrentPrefix, state.NamespaceStats)
	components.UpdateBigKeyTable(b.bigKeyTable, state.BigKeys)
	components.UpdateHotKeyTable(b.hotKeyTable, state.HotKeys)

	if state.KeyValue != nil {
		b.showKeyValue(state.KeyValue)
	}
}

func (b *BodyView) showKeyValue(info *models.KeyValueInfo) {
	b.keyValueView.Update(info)
	if b.showingValue {
		return
	}
	b.showingValue = true
	b.ContentFlex.Clear().AddItem(b.keyValueView.View, 0, 1, true)
	b.app.SetFocus(b.keyValueView.View)
}

func (b *BodyView) HandleKey(key rune, state *models.State) bool {
	if key == 't' {
		state.KeyValue = nil
		b.SetActiveView(b.nextTab())
		return true
	}
	if tab, ok := tabShortcuts[key]; ok {
		state.KeyValue = nil
		b.SetActiveView(tab)
		return true
	}

	switch b.activeView {
	case TabNamespace:
		if column, ok := namespaceSortColumns[key]; ok {
			state.NamespaceSort = column
			state.NamespaceStats.Sort(column)
			return true
		}
	case TabSlowLog:
		if column, ok := slowLogSortColumns[key]; ok {
			state.SlowLogs.Sort(column)
			return true
		}
	}
	return false
}

func (b *BodyView) ActiveView() Tab {
	return b.activeView
}

func (b *BodyView) IsShowingValue() bool {
	return b.showingValue
}

func (b *BodyView) SelectedIndex() int {
	row, _ := b.activeTable().GetSelection()
	return row - 1
}
