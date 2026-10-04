package ui

import (
	"fmt"
	"strings"

	"redscout/lib/scanner"
	"redscout/lib/ui/theme"
	"redscout/lib/ui/views"
	"redscout/lib/ui/views/components"
	"redscout/models"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type AppUI struct {
	config  *models.Config
	scanner *scanner.Scanner

	app      *tview.Application
	pages    *tview.Pages
	titleBar *views.TitleBar
	headers  *views.HeaderView
	body     *views.BodyView

	loadingTextView   *tview.TextView
	initialisedLayout bool
	helpVisible       bool
}

func NewAppUI(cfg models.Config) *AppUI {
	app := tview.NewApplication()

	ui := &AppUI{
		config:            &cfg,
		app:               app,
		pages:             tview.NewPages(),
		body:              views.NewBodyView(app),
		headers:           views.NewHeaderView(),
		titleBar:          views.NewTitleBar(),
		initialisedLayout: false,
		helpVisible:       false,
	}

	return ui
}

func (ui *AppUI) createDisclaimerScreen() {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBackgroundColor(theme.ColorBg)

	disclaimer := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[#f85149]DISCLAIMER[-]\n\n" +
			"[#8b949e]RedScout will run the 'MONITOR' command on your Redis instance.[-]\n" +
			"[#f0883e]This can impact Redis performance. Use with caution on production environments.[-]\n\n" +
			"[#c9d1d9]Do you want to continue?[-]\n\n" +
			"[#3fb950]Y[-]es / [#f85149]N[-]o")
	disclaimer.SetBorder(true)
	disclaimer.SetBorderColor(theme.ColorBorder)
	disclaimer.SetBackgroundColor(theme.ColorBg)
	disclaimer.SetBorderPadding(2, 2, 2, 2)

	flex.AddItem(disclaimer, 0, 1, false)
	ui.app.SetRoot(flex, true)

	ui.app.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		switch e.Rune() {
		case 'y', 'Y', '\r':
			ui.start()
			return nil
		case 'n', 'N', 'q', 'Q':
			ui.app.Stop()
			return nil
		}
		return e
	})
}

func (ui *AppUI) createErrorScreen(errorMsg string) {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBackgroundColor(theme.ColorBg)

	errorText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("[#f85149]ERROR[-]\n\n[#c9d1d9]%s[-]\n\n[#3fb950]R[-]etry / [#f85149]Q[-]uit", errorMsg))
	errorText.SetBorder(true)
	errorText.SetBorderColor(theme.ColorBorder)
	errorText.SetBackgroundColor(theme.ColorBg)
	errorText.SetBorderPadding(2, 2, 2, 2)

	flex.AddItem(errorText, 0, 1, false)
	ui.app.QueueUpdateDraw(func() {
		ui.app.SetRoot(flex, true)
	})

	ui.app.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		switch e.Rune() {
		case 'r', 'R', '\r':
			ui.start()
			return nil
		case 'q', 'Q':
			ui.app.Stop()
			return nil
		}
		return e
	})
}

func (ui *AppUI) start() {
	ui.createLoadingScreen()
	go func() {
		s, err := scanner.NewScanner(ui.config)
		if err != nil {
			ui.createErrorScreen(fmt.Sprintf("Error initializing scanner:\n%v", err))
			return
		}
		ui.scanner = s

		go s.Start()
		go ui.stateUpdateListener()
	}()
}

func (ui *AppUI) createLoadingScreen() {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBackgroundColor(theme.ColorBg)

	ui.loadingTextView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[#f0883e]Analysing Redis ⠋\n\n[#c9d1d9]Initializing...[-]")
	ui.loadingTextView.SetBorder(true)
	ui.loadingTextView.SetBorderColor(theme.ColorBorder)
	ui.loadingTextView.SetBackgroundColor(theme.ColorBg)
	ui.loadingTextView.SetBorderPadding(2, 2, 2, 2)

	flex.AddItem(ui.loadingTextView, 0, 1, false)
	ui.app.SetRoot(flex, true)

	spinner := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		i := 0

		for {
			select {
			case <-ticker.C:
				ui.app.QueueUpdateDraw(func() {
					var text string
					if ui.scanner == nil || ui.scanner.State == nil {
						text = fmt.Sprintf("[#f0883e]Analysing Redis %c\n\n[#c9d1d9]Initializing...[-]", spinner[i%len(spinner)])
					} else if ui.scanner.State.ScanComplete {
						ticker.Stop()
						return
					} else {
						var progressInfo string

						if ui.scanner.State.ScanProgress < 100 {
							scanBar := components.CreateProgressBar(ui.scanner.State.ScanProgress, 100, 40)
							progressInfo = fmt.Sprintf("\n\n[#8b949e]%s[-]\n%s\n[#c9d1d9]%d keys collected[-]", ui.scanner.State.Status, scanBar, ui.scanner.State.ScannedKeys)
						} else if ui.scanner.State.MonitorDurationTotal == 0 {
							progressInfo = "\n\n[#8b949e]Starting monitor...[-]"
						} else if ui.scanner.State.MonitorProgress < 100 {
							elapsed := time.Duration(float64(ui.scanner.State.MonitorDurationTotal) * ui.scanner.State.MonitorProgress / 100)
							monitorBar := components.CreateProgressBar(ui.scanner.State.MonitorProgress, 100, 40)
							progressInfo = fmt.Sprintf("\n\n[#8b949e]Monitor Progress:[-]\n%s\n[#c9d1d9]%v / %v[-]", monitorBar, elapsed.Round(time.Second), ui.scanner.State.MonitorDurationTotal)
						}

						text = fmt.Sprintf("[#f0883e]Analysing Redis %c[-]\n\n%s", spinner[i%len(spinner)], progressInfo)
					}
					ui.loadingTextView.SetText(text)
				})
				i++
			}
		}
	}()
}

func (ui *AppUI) createMainScreen() {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBackgroundColor(theme.ColorBg)

	flex.AddItem(ui.titleBar.View, 1, 0, false)
	flex.AddItem(ui.headers.HeaderFlex, 5, 0, false)

	ui.body.TabBar.SetBorder(false)
	ui.body.TabBar.SetBorderPadding(0, 0, 0, 0)
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

func (ui *AppUI) stateUpdateListener() {
	for update := range ui.scanner.State.Updates {
		if !update.ScanComplete {
			continue
		}

		ui.app.QueueUpdateDraw(func() {
			if !ui.initialisedLayout {
				ui.createMainScreen()
				ui.initialisedLayout = true
			}
			ui.update(update)
		})
	}
}

// fetchAndShowKeyValue fetches a key's value and shows it in the viewer.
// If key is empty, finds the first key matching the current namespace prefix.
func (ui *AppUI) fetchAndShowKeyValue(key string) {
	if key == "" {
		keys := ui.scanner.FindKeysWithPrefix(ui.scanner.State.CurrentPrefix, 1)
		if len(keys) == 0 {
			return
		}
		key = keys[0]
	}
	err := ui.scanner.FetchKeyValue(key)
	if err != nil {
		ui.scanner.State.KeyValue = &models.KeyValueInfo{
			Key:   key,
			Type:  "error",
			Value: fmt.Sprintf("Failed to fetch: %v", err),
		}
		ui.scanner.State.Updates <- ui.scanner.State
	}
}

func (ui *AppUI) Run() error {
	ui.createDisclaimerScreen()
	return ui.app.Run()
}

func (ui *AppUI) update(ctx *models.State) {
	ui.titleBar.Update(ctx)
	ui.headers.Update(ctx)
	ui.body.Update(ctx)
}

func (ui *AppUI) handleInput(e *tcell.EventKey) *tcell.EventKey {
	if e.Rune() == '?' {
		if ui.helpVisible {
			ui.pages.HidePage("help")
			ui.helpVisible = false
		} else {
			ui.pages.ShowPage("help")
			ui.helpVisible = true
		}
		return nil
	}
	if e.Key() == tcell.KeyEscape && ui.helpVisible {
		ui.pages.HidePage("help")
		ui.helpVisible = false
		return nil
	}
	if ui.helpVisible {
		return nil
	}

	changed := true

	switch e.Key() {
	case tcell.KeyEnter, tcell.KeyRight:
		if ui.body.IsShowingValue() {
			return nil
		}
		switch ui.body.ActiveView() {
		case "namespace":
			row, _ := ui.body.NamespaceTable().GetSelection()
			if row <= 0 || row > len(ui.scanner.State.NamespaceStats) {
				return nil
			}
			namespace := ui.scanner.State.NamespaceStats[row-1].Namespace
			ui.scanner.DrillDownNamespace(namespace)

			// If no sub-namespaces, find a matching key and fetch its value
			if len(ui.scanner.State.NamespaceStats) == 0 {
				go ui.fetchAndShowKeyValue("")
			}
			return nil
		case "bigkeys":
			row, _ := ui.body.BigKeyTable().GetSelection()
			if row <= 0 || row > len(ui.scanner.State.BigKeys) {
				return nil
			}
			key := ui.scanner.State.BigKeys[row-1].Key.String()
			go ui.fetchAndShowKeyValue(key)
			return nil
		case "hotkeys":
			row, _ := ui.body.HotKeyTable().GetSelection()
			if row <= 0 || row > len(ui.scanner.State.HotKeys) {
				return nil
			}
			key := ui.scanner.State.HotKeys[row-1].Key.String()
			go ui.fetchAndShowKeyValue(key)
			return nil
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyLeft:
		if ui.body.IsShowingValue() {
			ui.scanner.State.KeyValue = nil
			// Go back from key value view — level up past empty levels
			if ui.body.ActiveView() == "namespace" {
				for !ui.scanner.State.CurrentPrefix.IsEmpty() {
					ui.scanner.LevelUpNamespace()
					if len(ui.scanner.State.NamespaceStats) > 0 {
						break
					}
				}
			}
			ui.body.HideKeyValue()
			return nil
		}
		if ui.body.ActiveView() == "namespace" {
			ui.scanner.LevelUpNamespace()
			return nil
		}
	}

	switch e.Rune() {
	case '1', '2', '3', '4', '5', '6', '7', '8', 't', 'T', 'n', 'N', 'l', 'L', 'b', 'B', 'h', 'H':
		ui.body.HandleInput(e.Rune(), ui.scanner.State)
	case 'q', 'Q':
		ui.app.Stop()
		ui.scanner.Close()
		return nil
	case 's', 'S':
		go func() {
			err := ui.scanner.ScanMemory()
			if err == nil {
				_ = ui.scanner.ComputeNamespaceStats()
				_ = ui.scanner.ComputeBigKeysFromScanLog()
			}
		}()
	case 'm', 'M':
		go func() {
			err := ui.scanner.MonitorOps()
			if err == nil {
				_ = ui.scanner.ComputeNamespaceStats()
				_ = ui.scanner.ComputeHotKeysFromMonitorLog()
			}
		}()
	default:
		changed = false
	}
	if changed {
		ui.update(ui.scanner.State)
	}
	return e
}
