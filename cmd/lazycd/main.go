package main

import (
	"fmt"
	"log"
	"os"

	"lazycd/internal/core"
	"lazycd/internal/store"
	"lazycd/internal/ui"
)

func main() {
	// Initialize/Load state
	state, err := store.LoadState()
	if err != nil {
		fmt.Printf("Error loading state: %v\n", err)
		os.Exit(1)
	}

	// Always start in current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting cwd: %v\n", err)
		os.Exit(1)
	}
	state.LastDir = cwd

	// Initialize Job Manager
	jobMgr, err := core.NewJobManager()
	if err != nil {
		fmt.Printf("Error initializing job manager: %v\n", err)
		os.Exit(1)
	}

	// Initialize UI
	g := ui.NewGui(state, jobMgr)

	// Run UI
	if err := g.Run(); err != nil {
		log.Panicln(err)
	}

	// Save state on exit
	if err := state.Save(); err != nil {
		fmt.Printf("Error saving state: %v\n", err)
	}
}
