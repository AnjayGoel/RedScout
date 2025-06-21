package views

import (
	"github.com/rivo/tview"
	"redmon/lib/ui/views/components"
	"redmon/models"
)

type Tab string

const (
	TabNamespace   Tab = "namespace"
	TabSlowLog     Tab = "slowlog"
	TabSpecialKeys Tab = "specialkeys"
)

type BodyView struct {
	Shortcuts       *tview.TextView
	ContentFlex     *tview.Flex
	namespaceTable  *tview.Table
	slowLogTable    *tview.Table
	activeView      Tab
	TabBar          *tview.TextView
	app             *tview.Application
	specialKeysView *components.SpecialKeysView
}

func NewBodyView(app *tview.Application) *BodyView {
	view := &BodyView{
		app:             app,
		Shortcuts:       buildShortcuts(),
		ContentFlex:     buildContentFlex(),
		namespaceTable:  components.BuildStatsTable(),
		slowLogTable:    components.BuildSlowLogTable(),
		activeView:      "namespace",
		TabBar:          buildTabBar(),
		specialKeysView: components.NewSpecialKeysView(app),
	}
	view.SetActiveView("namespace")
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

func buildShortcuts() *tview.TextView {
	return tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)
}

func buildContentFlex() *tview.Flex {
	return tview.NewFlex().SetDirection(tview.FlexColumn)
}

func buildTabBar() *tview.TextView {
	tabBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	return tabBar
}

func (b *BodyView) SetActiveView(view Tab) {
	b.activeView = view

	switch view {
	case TabNamespace:
		b.ContentFlex.Clear().AddItem(b.namespaceTable, 0, 2, true)
		b.Shortcuts.SetText(components.StatsHeader)
		b.TabBar.SetText(
			`[::b][white:teal]Namespace [white::-]` +
				`[white:black][-:-]` +
				`[white] Slow Log [-]` +
				`[white:black][-:-]` +
				`[white] Big Keys/Hot Keys [-]`,
		)
		b.namespaceTable.Select(1, 0)
		b.app.SetFocus(b.namespaceTable)
	case TabSlowLog:
		b.ContentFlex.Clear().AddItem(b.slowLogTable, 0, 2, true)
		b.Shortcuts.SetText(components.SlowLogHeader)
		b.TabBar.SetText(
			`[white]Namespace [-]` +
				`[white:black][-:-]` +
				`[::b][white:teal] Slow Log [white::-]` +
				`[white:black][-:-]` +
				`[white] Big Keys/Hot Keys [-]`,
		)
		b.slowLogTable.Select(1, 0)
		b.app.SetFocus(b.slowLogTable)
	case TabSpecialKeys:
		b.ContentFlex.Clear().AddItem(b.specialKeysView.Flex, 0, 2, true)
		b.Shortcuts.SetText(components.SpecialKeysShortcutsText)
		b.TabBar.SetText(
			`[white]Namespace [-]` +
				`[white:black][-:-]` +
				`[white] Slow Log [-]` +
				`[white:black][-:-]` +
				`[::b][white:teal] Big Keys/Hot Keys [white::-]`,
		)
		b.specialKeysView.Focus()
	}
}

func (b *BodyView) ToggleView() {
	switch b.activeView {
	case TabNamespace:
		b.SetActiveView(TabSlowLog)
	case TabSlowLog:
		b.SetActiveView(TabSpecialKeys)
	default:
		b.SetActiveView(TabNamespace)
	}
}

func (b *BodyView) Update(data *models.State) {
	components.PopulateSlowLogTable(b.slowLogTable, data.SlowLogs)
	components.PopulateStatsTable(b.namespaceTable, data.NamespaceStats)
	b.specialKeysView.Update(data)
}

func (b *BodyView) HandleInput(inp rune, state *models.State) {
	if inp == 'T' || inp == 't' {
		b.ToggleView()
		return
	}
	if inp == 'B' || inp == 'b' {
		b.SetActiveView(TabSpecialKeys)
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
	return b.namespaceTable
}
