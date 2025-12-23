package ui

import (
	"fmt"

	"lazycd/internal/core"
	"lazycd/internal/store"

	"github.com/awesome-gocui/gocui"
)

type Gui struct {
	g     *gocui.Gui
	State *store.State
	JobMgr *core.JobManager

	Browser *Browser
	Shelf   *Shelf
}

func NewGui(state *store.State, jobMgr *core.JobManager) *Gui {
	return &Gui{
		State:  state,
		JobMgr: jobMgr,
	}
}

func (gui *Gui) Run() error {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		return err
	}
	defer g.Close()

	gui.g = g
	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen
	g.InputEsc = true

	gui.Browser = NewBrowser(gui)
	gui.Shelf = NewShelf(gui)

	g.SetManagerFunc(gui.layout)

	if err := gui.keybindings(); err != nil {
		return err
	}

	if err := gui.Browser.Init(); err != nil {
		return err
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

func (gui *Gui) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Browser view (Left)
	if v, err := g.SetView("browser", 0, 0, maxX/2-1, maxY-2, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Browser "
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		
		gui.Browser.Draw(v)
		
		if _, err := g.SetCurrentView("browser"); err != nil {
			return err
		}
	}

	// Shelf view (Right)
	if v, err := g.SetView("shelf", maxX/2, 0, maxX-1, maxY-2, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = fmt.Sprintf(" Shelf (%d) ", len(gui.State.ShelfItems))
	}

	// Status view (Bottom)
	if v, err := g.SetView("status", 0, maxY-2, maxX-1, maxY, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.BgColor = gocui.ColorBlue
		v.FgColor = gocui.ColorWhite
	}
	
	gui.updateStatus()

	return nil
}

func (gui *Gui) keybindings() error {
	if err := gui.g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, failQuit); err != nil {
		return err
	}
	if err := gui.g.SetKeybinding("", 'q', gocui.ModNone, failQuit); err != nil {
		return err
	}
	
	// Tab to switch views
	if err := gui.g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, gui.nextView); err != nil {
		return err
	}

	if err := gui.g.SetKeybinding("", 'u', gocui.ModNone, gui.undoLastJob); err != nil {
		return err
	}
	
	if err := gui.Browser.Keybindings(); err != nil {
		return err
	}
	
	if err := gui.Shelf.Keybindings(); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) nextView(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == "browser" {
		_, err := g.SetCurrentView("shelf")
		return err
	}
	_, err := g.SetCurrentView("browser")
	return err
}

func failQuit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (gui *Gui) updateStatus() {
	v, err := gui.g.View("status")
	if err != nil {
		return
	}
	v.Clear()
	
	cwd := gui.State.LastDir
	target := gui.State.TargetDir
	if target == "" {
		target = "(none)"
	}
	
	fmt.Fprintf(v, " CWD: %s | Target: %s | Tab: Switch View | q: Quit", cwd, target)
}
