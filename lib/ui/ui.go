package ui

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"redscout/lib/scanner"
	"redscout/lib/ui/theme"
	"redscout/lib/ui/views"
	"redscout/lib/ui/views/components"
	"redscout/models"
)

const disclaimerText = "[#f85149]DISCLAIMER[-]\n\n" +
	"[#8b949e]RedScout will run the 'MONITOR' command on your Redis instance.[-]\n" +
	"[#f0883e]This can impact Redis performance. Use with caution on production environments.[-]\n\n" +
	"[#c9d1d9]Do you want to continue?[-]\n\n" +
	"[#3fb950]Y[-]es / [#f85149]N[-]o"

var spinner = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

type AppUI struct {
	config  *models.Config
	scanner *scanner.Scanner

	app      *tview.Application
	pages    *tview.Pages
	titleBar *views.TitleBar
	headers  *views.HeaderView
	body     *views.BodyView

	loading     *tview.TextView
	loadingDone chan struct{}
	mainShown   bool
	helpVisible bool
}

func NewAppUI(cfg models.Config) *AppUI {
	app := tview.NewApplication()
	return &AppUI{
		config:   &cfg,
		app:      app,
		pages:    tview.NewPages(),
		titleBar: views.NewTitleBar(),
		headers:  views.NewHeaderView(),
		body:     views.NewBodyView(app),
	}
}

func (ui *AppUI) Run() error {
	ui.showDisclaimer()
	err := ui.app.Run()
	if ui.scanner != nil {
		ui.scanner.Close()
	}
	return err
}

func newMessageView(text string) *tview.TextView {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(text)
	view.SetBorder(true)
	view.SetBorderColor(theme.ColorBorder)
	view.SetBackgroundColor(theme.ColorBg)
	view.SetBorderPadding(2, 2, 2, 2)
	return view
}

func (ui *AppUI) showMessage(text string, onKey func(r rune)) *tview.TextView {
	view := newMessageView(text)
	ui.app.SetRoot(view, true)
	ui.app.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		onKey(unicode.ToLower(e.Rune()))
		return e
	})
	return view
}

func (ui *AppUI) showDisclaimer() {
	ui.showMessage(disclaimerText, func(r rune) {
		switch r {
		case 'y', '\r':
			ui.start()
		case 'n', 'q':
			ui.app.Stop()
		}
	})
}

func (ui *AppUI) showError(msg string) {
	ui.stopLoading()
	text := fmt.Sprintf("[#f85149]ERROR[-]\n\n[#c9d1d9]%s[-]\n\n[#3fb950]R[-]etry / [#f85149]Q[-]uit", tview.Escape(msg))
	ui.showMessage(text, func(r rune) {
		switch r {
		case 'r', '\r':
			ui.start()
		case 'q':
			ui.app.Stop()
		}
	})
}

func (ui *AppUI) start() {
	ui.loading = ui.showMessage(ui.loadingText(0), func(r rune) {
		if r == 'q' {
			ui.app.Stop()
		}
	})
	ui.loadingDone = make(chan struct{})
	go ui.animateLoading(ui.loadingDone)

	go func() {
		s, err := scanner.NewScanner(ui.config)
		if err != nil {
			ui.app.QueueUpdateDraw(func() {
				ui.showError(fmt.Sprintf("Error initializing scanner:\n%v", err))
			})
			return
		}
		ui.scanner = s
		go s.Start()
		ui.listen(s)
	}()
}

func (ui *AppUI) stopLoading() {
	if ui.loadingDone != nil {
		close(ui.loadingDone)
		ui.loadingDone = nil
	}
}

func (ui *AppUI) animateLoading(done <-chan struct{}) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for frame := 1; ; frame++ {
		select {
		case <-done:
			return
		case <-ticker.C:
			ui.app.QueueUpdateDraw(func() {
				ui.loading.SetText(ui.loadingText(frame))
			})
		}
	}
}

func (ui *AppUI) loadingText(frame int) string {
	title := fmt.Sprintf("[#f0883e]Analysing Redis %c[-]\n\n", spinner[frame%len(spinner)])
	if ui.scanner == nil {
		return title + "[#c9d1d9]Initializing...[-]"
	}

	state := ui.scanner.State
	status := tview.Escape(state.Status)
	switch {
	case state.ScanProgress < 100:
		bar := components.CreateProgressBar(state.ScanProgress, 100, 40)
		return title + fmt.Sprintf("[#8b949e]%s[-]\n%s\n[#c9d1d9]%d keys collected[-]", status, bar, state.ScannedKeys)
	case state.MonitorProgress > 0 && state.MonitorProgress < 100:
		total := ui.config.MonitorDuration
		elapsed := time.Duration(float64(total) * state.MonitorProgress / 100)
		bar := components.CreateProgressBar(state.MonitorProgress, 100, 40)
		return title + fmt.Sprintf("[#8b949e]Monitor Progress:[-]\n%s\n[#c9d1d9]%v / %v[-]", bar, elapsed.Round(time.Second), total)
	default:
		return title + fmt.Sprintf("[#8b949e]%s[-]", status)
	}
}

func (ui *AppUI) showMain() {
	ui.stopLoading()
	ui.mainShown = true

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBackgroundColor(theme.ColorBg)
	flex.AddItem(ui.titleBar.View, 1, 0, false)
	flex.AddItem(ui.headers.HeaderFlex, 5, 0, false)
	flex.AddItem(ui.body.TabBar, 1, 0, false)

	// Thin separator line below tabs
	tabSep := tview.NewTextView().SetDynamicColors(true)
	tabSep.SetBackgroundColor(theme.ColorBg)
	tabSep.SetText("[#21262d]" + strings.Repeat("─", 300) + "[-]")
	flex.AddItem(tabSep, 1, 0, false)

	flex.AddItem(ui.body.ContentFlex, 0, 1, true)

	ui.pages.AddPage("main", flex, true, true)
	ui.pages.AddPage("help", views.NewHelpOverlay(), true, false)

	ui.app.SetInputCapture(ui.handleInput)
	ui.app.SetRoot(ui.pages, true)
}

func (ui *AppUI) listen(s *scanner.Scanner) {
	for state := range s.State.Updates {
		if !state.ScanComplete {
			continue
		}
		ui.app.QueueUpdateDraw(func() {
			if !ui.mainShown {
				ui.showMain()
			}
			ui.refresh()
		})
	}
}

func (ui *AppUI) refresh() {
	state := ui.scanner.State
	ui.titleBar.Update(state)
	ui.headers.Update(state)
	ui.body.Update(state)
}

func (ui *AppUI) handleInput(e *tcell.EventKey) *tcell.EventKey {
	if e.Rune() == '?' || (e.Key() == tcell.KeyEscape && ui.helpVisible) {
		ui.toggleHelp()
		return nil
	}
	if ui.helpVisible {
		return nil
	}

	switch e.Key() {
	case tcell.KeyEnter, tcell.KeyRight:
		ui.openSelection()
		return nil
	case tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyLeft:
		ui.goBack()
		return nil
	}

	switch key := unicode.ToLower(e.Rune()); key {
	case 'q':
		ui.app.Stop()
	case 's':
		go ui.scanner.Rescan()
	case 'm':
		go ui.scanner.Remonitor()
	default:
		if !ui.body.HandleKey(key, ui.scanner.State) {
			return e
		}
		ui.refresh()
	}
	return nil
}

func (ui *AppUI) toggleHelp() {
	ui.helpVisible = !ui.helpVisible
	if ui.helpVisible {
		ui.pages.ShowPage("help")
	} else {
		ui.pages.HidePage("help")
	}
}

func (ui *AppUI) openSelection() {
	if ui.body.IsShowingValue() {
		return
	}

	state := ui.scanner.State
	i := ui.body.SelectedIndex()
	if i < 0 {
		return
	}

	switch ui.body.ActiveView() {
	case views.TabNamespace:
		if i < len(state.NamespaceStats) {
			ui.scanner.DrillDownNamespace(state.NamespaceStats[i].Namespace)
		}
	case views.TabBigKeys:
		if i < len(state.BigKeys) {
			go ui.scanner.FetchKeyValue(state.BigKeys[i].Key)
		}
	case views.TabHotKeys:
		if i < len(state.HotKeys) {
			go ui.scanner.FetchKeyValue(state.HotKeys[i].Key)
		}
	}
}

func (ui *AppUI) goBack() {
	if ui.body.IsShowingValue() {
		ui.scanner.State.KeyValue = nil
		ui.body.HideKeyValue()
	}
	if ui.body.ActiveView() == views.TabNamespace {
		ui.scanner.LevelUpNamespace()
	}
}
