package main

import (
	"log"
	"os"
	"redscout/lib"
	"redscout/lib/ui"
)

func main() {
	log.SetFlags(0)
	for _, arg := range os.Args {
		if arg == "-h" || arg == "--help" {
			os.Setenv("TVIEW_DISABLE", "1")
			break
		}
	}

	cfg := lib.ParseFlags()

	// Disable Tview UI if showing help
	if os.Getenv("TVIEW_DISABLE") == "1" {
		return
	}

	app := ui.NewAppUI(cfg)
	if err := app.Run(); err != nil {
		log.Printf("Error running application: %v\n", err)
	}
}
