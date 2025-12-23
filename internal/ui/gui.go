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
	if err := gui.g.SetKeybinding("", '?', gocui.ModNone, gui.toggleHelp); err != nil {
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

func (gui *Gui) toggleHelp(g *gocui.Gui, v *gocui.View) error {
	_, err := g.View("help")
	if err == nil {
		if err := g.DeleteView("help"); err != nil {
			return err
		}
		// Restore focus to browser (default). Ideally we track previous view but browser is fine for MVP.
		_, err := g.SetCurrentView("browser")
		return err
	}

	maxX, maxY := g.Size()
	
	// Centered modal
	v, err = g.SetView("help", maxX/6, maxY/6, maxX*5/6, maxY*5/6, 0)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	v.Title = " Help (Close: ?) "
	
	fmt.Fprintln(v, "Global Keys:")
	fmt.Fprintln(v, "  Tab: Switch View (Browser <-> Shelf)")
	fmt.Fprintln(v, "  ?: Toggle Help")
	fmt.Fprintln(v, "  u: Undo last job")
	fmt.Fprintln(v, "  q: Quit")
	fmt.Fprintln(v, "")
	fmt.Fprintln(v, "Browser Keys:")
	fmt.Fprintln(v, "  j/k/Down/Up: Navigation")
	fmt.Fprintln(v, "  l/Right: Enter Directory")
	fmt.Fprintln(v, "  h/Left: Parent Directory")
	fmt.Fprintln(v, "  Space: Multi-select")
	fmt.Fprintln(v, "  a: Add to Shelf")
	fmt.Fprintln(v, "  t: Set Target")
	fmt.Fprintln(v, "  .: Toggle Hidden Files")
	fmt.Fprintln(v, "")
	fmt.Fprintln(v, "Shelf Keys:")
	fmt.Fprintln(v, "  y: Set mode to Copy")
	fmt.Fprintln(v, "  x: Set mode to Move")
	fmt.Fprintln(v, "  r: Remove from Shelf")
	fmt.Fprintln(v, "  d: Delete items")
	fmt.Fprintln(v, "  p: Put items to Target")

	if _, err := g.SetCurrentView("help"); err != nil {
		return err
	}
	
	return nil
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
	
	fmt.Fprintf(v, " CWD: %s | Target: %s | Tab: Switch View | ?: Help | q: Quit", cwd, target)
}
