package ui

import (
	"fmt"
	"os"

	"lazycd/internal/store"

	"github.com/awesome-gocui/gocui"
)

type Shelf struct {
	gui *Gui
	
	// Selection state
	selected map[string]struct{} // Set of IDs
}

func NewShelf(gui *Gui) *Shelf {
	return &Shelf{
		gui:      gui,
		selected: make(map[string]struct{}),
	}
}

func (s *Shelf) Keybindings() error {
	if err := s.gui.g.SetKeybinding("shelf", 'j', gocui.ModNone, s.cursorDown); err != nil {
		return err
	}
	if err := s.gui.g.SetKeybinding("shelf", gocui.KeyArrowDown, gocui.ModNone, s.cursorDown); err != nil {
		return err
	}
	if err := s.gui.g.SetKeybinding("shelf", 'k', gocui.ModNone, s.cursorUp); err != nil {
		return err
	}
	if err := s.gui.g.SetKeybinding("shelf", gocui.KeyArrowUp, gocui.ModNone, s.cursorUp); err != nil {
		return err
	}
	if err := s.gui.g.SetKeybinding("shelf", gocui.KeySpace, gocui.ModNone, s.toggleSelect); err != nil {
		return err
	}
	if err := s.gui.g.SetKeybinding("shelf", 'r', gocui.ModNone, s.remove); err != nil {
		return err
	}
	return nil
}

func (s *Shelf) cursors() (int, int) {
	v, err := s.gui.g.View("shelf")
	if err != nil {
		return 0, 0
	}
	return v.Cursor()
}

func (s *Shelf) cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if cy < len(s.gui.State.ShelfItems)-1 {
			if err := v.SetCursor(cx, cy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Shelf) cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if cy > 0 {
			if err := v.SetCursor(cx, cy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Shelf) toggleSelect(g *gocui.Gui, v *gocui.View) error {
	item := s.currentItem(v)
	if item == nil {
		return nil
	}
	
	if _, ok := s.selected[item.ID]; ok {
		delete(s.selected, item.ID)
	} else {
		s.selected[item.ID] = struct{}{}
	}
	
	s.Update()
	return nil
}

func (s *Shelf) remove(g *gocui.Gui, v *gocui.View) error {
	// Remove selected items
	newItems := []store.ShelfItem{}
	
	// If nothing selected, remove current item
	if len(s.selected) == 0 {
		currentItem := s.currentItem(v)
		if currentItem == nil {
			return nil
		}
		// Create temp map for removal logic
		targets := map[string]struct{}{currentItem.ID: {}}
		
		for _, item := range s.gui.State.ShelfItems {
			if _, ok := targets[item.ID]; !ok {
				newItems = append(newItems, item)
			}
		}
	} else {
		for _, item := range s.gui.State.ShelfItems {
			if _, ok := s.selected[item.ID]; !ok {
				newItems = append(newItems, item)
			}
		}
		// Clear selection
		s.selected = make(map[string]struct{})
	}
	
	s.gui.State.ShelfItems = newItems
	
	// Fix cursor if it's out of bounds
	cx, cy := v.Cursor()
	if cy >= len(newItems) && cy > 0 {
		v.SetCursor(cx, cy-1)
	}
	
	s.Update()
	s.gui.updateStatus()
	return nil
}

func (s *Shelf) currentItem(v *gocui.View) *store.ShelfItem {
	_, cy := v.Cursor()
	if cy >= 0 && cy < len(s.gui.State.ShelfItems) {
		return &s.gui.State.ShelfItems[cy]
	}
	return nil
}

func (s *Shelf) Update() {
	v, err := s.gui.g.View("shelf")
	if err != nil {
		return // View might not be ready
	}
	s.gui.g.Update(func(g *gocui.Gui) error {
		v.Clear()
		v.Title = fmt.Sprintf(" Shelf (%d) ", len(s.gui.State.ShelfItems))
		
		for _, item := range s.gui.State.ShelfItems {
			mark := " "
			if _, ok := s.selected[item.ID]; ok {
				mark = "*"
			}
			
			// Check stale
			staleMark := ""
			if _, err := os.Stat(item.AbsPath); os.IsNotExist(err) {
				staleMark = " (!)"
			}
			
			fmt.Fprintf(v, "%s [%s] %s (%s)%s\n", mark, item.Type, item.AbsPath, item.OpMode, staleMark)
		}
		return nil
	})
}
