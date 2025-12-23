package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"lazycd/internal/fs"
	"lazycd/internal/store"

	"github.com/awesome-gocui/gocui"
)

type Browser struct {
	gui   *Gui
	items []fs.FileItem
	
	// Selection state
	selected map[string]struct{} // Set of absolute paths
}

func NewBrowser(gui *Gui) *Browser {
	return &Browser{
		gui:      gui,
		selected: make(map[string]struct{}),
	}
}

func (b *Browser) Init() error {
	if b.gui.State.LastDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		b.gui.State.LastDir = cwd
	}
	
	// Only fetch data, do not try to update view as it doesn't exist yet
	return b.Fetch()
}

func (b *Browser) Fetch() error {
	var err error
	b.items, err = fs.ListDir(b.gui.State.LastDir)
	return err
}

func (b *Browser) Refresh() error {
	if err := b.Fetch(); err != nil {
		return err
	}
	
	b.UpdateView()
	b.gui.updateStatus()
	return nil
}

func (b *Browser) Draw(v *gocui.View) {
	v.Clear()
	v.Title = fmt.Sprintf(" Browser: %s ", b.gui.State.LastDir)
	
	for _, item := range b.items {
		mark := " "
		if _, ok := b.selected[item.Path]; ok {
			mark = "*"
		}
		
		suffix := ""
		if item.IsDir {
			suffix = "/"
		}
		
		fmt.Fprintf(v, "%s %s%s\n", mark, item.Name, suffix)
	}
}

func (b *Browser) UpdateView() {
	b.gui.g.Update(func(g *gocui.Gui) error {
		v, err := g.View("browser")
		if err != nil {
			return nil
		}
		b.Draw(v)
		return nil
	})
}

func (b *Browser) Keybindings() error {
	if err := b.gui.g.SetKeybinding("browser", 'j', gocui.ModNone, b.cursorDown); err != nil {
		return err
	}
	if err := b.gui.g.SetKeybinding("browser", gocui.KeyArrowDown, gocui.ModNone, b.cursorDown); err != nil {
		return err
	}
	if err := b.gui.g.SetKeybinding("browser", 'k', gocui.ModNone, b.cursorUp); err != nil {
		return err
	}
	if err := b.gui.g.SetKeybinding("browser", gocui.KeyArrowUp, gocui.ModNone, b.cursorUp); err != nil {
		return err
	}
	if err := b.gui.g.SetKeybinding("browser", 'l', gocui.ModNone, b.enterDir); err != nil {
		return err
	}
	if err := b.gui.g.SetKeybinding("browser", gocui.KeyArrowRight, gocui.ModNone, b.enterDir); err != nil {
		return err
	}
	if err := b.gui.g.SetKeybinding("browser", 'h', gocui.ModNone, b.parentDir); err != nil {
		return err
	}
	if err := b.gui.g.SetKeybinding("browser", gocui.KeyArrowLeft, gocui.ModNone, b.parentDir); err != nil {
		return err
	}
	if err := b.gui.g.SetKeybinding("browser", gocui.KeySpace, gocui.ModNone, b.toggleSelect); err != nil {
		return err
	}
	if err := b.gui.g.SetKeybinding("browser", 'a', gocui.ModNone, b.addToShelf); err != nil {
		return err
	}
	if err := b.gui.g.SetKeybinding("browser", 't', gocui.ModNone, b.setTarget); err != nil {
		return err
	}
	return nil
}

func (b *Browser) cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		ox, oy := v.Origin()
		
		// If we are selecting an item beyond the visible bottom
		if cy+oy < len(b.items)-1 {
			// If we are at the bottom of the view, move origin down
			_, vy := v.Size()
			if cy >= vy-1 {
				if err := v.SetOrigin(ox, oy+1); err != nil {
					return err
				}
			} else {
				if err := v.SetCursor(cx, cy+1); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (b *Browser) cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		
		if cy == 0 {
			if oy > 0 {
				if err := v.SetOrigin(ox, oy-1); err != nil {
					return err
				}
			}
		} else {
			if err := v.SetCursor(cx, cy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Browser) enterDir(g *gocui.Gui, v *gocui.View) error {
	item := b.currentItem(v)
	if item == nil {
		return nil
	}
	
	if item.IsDir {
		b.gui.State.LastDir = item.Path
		if err := b.Refresh(); err != nil {
			return err
		}
		v.SetCursor(0, 0)
	}
	return nil
}

func (b *Browser) parentDir(g *gocui.Gui, v *gocui.View) error {
	parent := filepath.Dir(b.gui.State.LastDir)
	if parent != b.gui.State.LastDir {
		b.gui.State.LastDir = parent
		if err := b.Refresh(); err != nil {
			return err
		}
		v.SetCursor(0, 0)
	}
	return nil
}

func (b *Browser) toggleSelect(g *gocui.Gui, v *gocui.View) error {
	item := b.currentItem(v)
	if item == nil {
		return nil
	}
	
	if _, ok := b.selected[item.Path]; ok {
		delete(b.selected, item.Path)
	} else {
		b.selected[item.Path] = struct{}{}
	}
	
	b.UpdateView()
	// Restore cursor position trick or just full refresh?
	// updateView rewrites content, so cursor might be fine conceptually but
	// gocui view buffer is replaced. Gocui usually keeps cursor unless reset.
	return nil
}

func (b *Browser) addToShelf(g *gocui.Gui, v *gocui.View) error {
	// Add selected items
	count := 0
	for path := range b.selected {
		// Check duplicates
		exists := false
		for _, sVal := range b.gui.State.ShelfItems {
			if sVal.AbsPath == path {
				exists = true
				break
			}
		}
		if !exists {
			info, err := os.Stat(path)
			isDir := false
			if err == nil {
				isDir = info.IsDir()
			}
			b.gui.State.ShelfItems = append(b.gui.State.ShelfItems, store.NewShelfItem(path, isDir))
			count++
		}
	}
	
	// IF no selection, add current item
	if len(b.selected) == 0 {
		item := b.currentItem(v)
		if item != nil {
			exists := false
			for _, sVal := range b.gui.State.ShelfItems {
				if sVal.AbsPath == item.Path {
					exists = true
					break
				}
			}
			if !exists {
				b.gui.State.ShelfItems = append(b.gui.State.ShelfItems, store.NewShelfItem(item.Path, item.IsDir))
				count++
			}
		}
	}
	
	// Clear selection after add? Spec is vague, but usually yes.
	// But "Shelf에서 다중 선택" implies browser selection might stay?
	// Let's keep it for now.
	
	b.gui.Shelf.Update()
	b.gui.updateStatus()
	return nil
}

func (b *Browser) setTarget(g *gocui.Gui, v *gocui.View) error {
	// Spec: t = 현재 브라우저 디렉토리를 target으로 설정
	b.gui.State.TargetDir = b.gui.State.LastDir
	b.gui.updateStatus()
	return nil
}

func (b *Browser) currentItem(v *gocui.View) *fs.FileItem {
	_, cy := v.Cursor()
	_, oy := v.Origin()
	index := cy + oy
	if index >= 0 && index < len(b.items) {
		return &b.items[index]
	}
	return nil
}
