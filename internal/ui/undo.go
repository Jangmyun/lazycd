package ui

import (
	"github.com/awesome-gocui/gocui"
)

func (gui *Gui) undoLastJob(g *gocui.Gui, v *gocui.View) error {
	jobs, err := gui.JobMgr.GetRecentJobs(1)
	if err != nil {
		return nil
	}
	
	if len(jobs) == 0 {
		return nil // Nothing to undo
	}
	
	lastJob := jobs[0]
	if err := gui.JobMgr.Undo(lastJob); err != nil {
		// Show error log?
	}
	
	// Refresh views
	if gui.Browser != nil {
		gui.Browser.Refresh()
	}
	gui.updateStatus()
	
	return nil
}
