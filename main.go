package main

import (
	"log"

	"redscout/lib"
	"redscout/lib/ui"
)

func main() {
	log.SetFlags(0)

	cfg, err := lib.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	if err := ui.NewAppUI(cfg).Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
