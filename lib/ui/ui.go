package ui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"log"
	"redscout/lib/scanner"
	"redscout/lib/ui/views"
	"redscout/models"
	"time"
)

type AppUI struct {
	//Config
	config *models.Config

	//TODO: Decouple UI from Scanner
	scanner *scanner.Scanner

	//UI Components
	app     *tview.Application
	headers *views.HeaderView
	body    *views.BodyView

	// Loading screen components
	loadingTextView *tview.TextView

	initialisedLayout bool
}

func NewAppUI(cfg models.Config) *AppUI {
	app := tview.NewApplication()

	ui := &AppUI{
		config:            &cfg,
		app:               app,
		body:              views.NewBodyView(app),
		headers:           views.NewHeaderView(),
		initialisedLayout: false,
	}

	return ui
}

func (ui *AppUI) createDisclaimerScreen() {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	disclaimer := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[red]DISCLAIMER[-]\n\n" +
			"[yellow]RedScout will run the 'MONITOR' command on your Redis instance.[-]\n" +
			"[yellow]This can impact Redis performance. Use with caution on production environments.[-]\n\n" +
			"[white]Do you want to continue?[white]\n\n" +
			"[green]Y[-]es / [red]N[-]o")
	disclaimer.SetBorder(true)
	disclaimer.SetBorderPadding(2, 2, 2, 2)

	flex.AddItem(disclaimer, 0, 1, false)
	ui.app.SetRoot(flex, true)

	// Set up input capture for disclaimer
	ui.app.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		switch e.Rune() {
		case 'y', 'Y':
			ui.start()
			return nil
		case 'n', 'N':
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
			ui.app.Stop()
			log.Fatalf("Error initializing scanner: %v", err)
		}
		ui.scanner = s

		go s.Start()
		go ui.stateUpdateListener()
	}()
}

func (ui *AppUI) createLoadingScreen() {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	ui.loadingTextView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]Loading RedScout ⠋\n\n[white]Initializing...[-]")
	ui.loadingTextView.SetBorder(true)
	ui.loadingTextView.SetBorderPadding(2, 2, 2, 2)

	flex.AddItem(ui.loadingTextView, 0, 1, false)
	ui.app.SetRoot(flex, true)

	// Spinner animation
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
						text = fmt.Sprintf("[yellow]Loading RedScout %c\n\n[white]Initializing...[-]", spinner[i%len(spinner)])
					} else if ui.scanner.State.ScanComplete {
						ticker.Stop()
						return
					} else {
						text = fmt.Sprintf("[yellow]Loading RedScout %c\n\n[white]%s[-]", spinner[i%len(spinner)], ui.scanner.State.Status)
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
	flex.Clear()
	flex.AddItem(ui.headers.HeaderFlex, 6, 0, false)
	flex.AddItem(ui.body.TabBar, 1, 0, false)
	flex.AddItem(ui.body.ContentFlex, 0, 1, true)
	flex.AddItem(ui.body.Shortcuts, 1, 0, false)
	ui.app.SetInputCapture(ui.handleInput)
	ui.app.SetRoot(flex, true)
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

func (ui *AppUI) Run() error {
	ui.createDisclaimerScreen()
	return ui.app.Run()
}

func (ui *AppUI) update(ctx *models.State) {
	ui.headers.Update(ctx)
	ui.body.Update(ctx)
}

func (ui *AppUI) handleInput(e *tcell.EventKey) *tcell.EventKey {
	changed := true

	switch e.Key() {
	case tcell.KeyEnter, tcell.KeyRight:
		if ui.body.ActiveView() == "namespace" {
			row, _ := ui.body.NamespaceTable().GetSelection()
			if row <= 0 || row > len(ui.scanner.State.NamespaceStats) {
				return nil
			}
			namespace := ui.scanner.State.NamespaceStats[row-1].Namespace

			ui.scanner.DrillDownNamespace(namespace)

			return nil
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyLeft:
		if ui.body.ActiveView() == "namespace" {
			ui.scanner.LevelUpNamespace()
			return nil
		}
	}

	switch e.Rune() {
	case '1', '2', '3', '4', '5', '6', '7', '8', 't', 'T':
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
			}
		}()
	case 'm', 'M':
		go func() {
			err := ui.scanner.MonitorOps()
			if err == nil {
				_ = ui.scanner.ComputeNamespaceStats()
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
