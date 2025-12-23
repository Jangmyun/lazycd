package ui

import (
	"path/filepath"

	"lazycd/internal/core"
	"lazycd/internal/store"

	"github.com/awesome-gocui/gocui"
)

func (s *Shelf) setModeCopy(g *gocui.Gui, v *gocui.View) error {
	s.setMode(store.OpCopy)
	return nil
}

func (s *Shelf) setModeMove(g *gocui.Gui, v *gocui.View) error {
	s.setMode(store.OpMove)
	return nil
}

func (s *Shelf) setMode(mode store.OpMode) {
	// If selection, apply to selected. Else apply to all? Or current?
	// Spec: "Shelf에서 다중 선택 -> Copy/Move... "
	// H1: "shelf 선택 항목에 적용"
	
	if len(s.selected) > 0 {
		for i, item := range s.gui.State.ShelfItems {
			if _, ok := s.selected[item.ID]; ok {
				s.gui.State.ShelfItems[i].OpMode = mode
			}
		}
	} else {
		// If nothing selected, apply to current item
		v, _ := s.gui.g.View("shelf")
		if v != nil {
			item := s.currentItem(v)
			if item != nil {
				// Find index
				for i, it := range s.gui.State.ShelfItems {
					if it.ID == item.ID {
						s.gui.State.ShelfItems[i].OpMode = mode
						break
					}
				}
			}
		}
	}
	s.Update()
}

func (s *Shelf) executeDelete(g *gocui.Gui, v *gocui.View) error {
	// Items to delete
	var targets []int
	if len(s.selected) > 0 {
		for i, item := range s.gui.State.ShelfItems {
			if _, ok := s.selected[item.ID]; ok {
				targets = append(targets, i)
			}
		}
	} else {
		// Current item
		item := s.currentItem(v)
		if item != nil {
			for i, it := range s.gui.State.ShelfItems {
				if it.ID == item.ID {
					targets = append(targets, i)
					break
				}
			}
		}
	}
	
	if len(targets) == 0 {
		return nil
	}
	
	// Create Job
	job := s.gui.JobMgr.CreateJob(core.JobDelete)
	
	// Execute
	for _, idx := range targets {
		item := s.gui.State.ShelfItems[idx]
		
		trashPath, err := core.DeleteToTrash(item.AbsPath, job.ID)
		
		jobItem := core.JobItem{
			Src:    item.AbsPath,
			Op:     "delete",
			Status: core.StatusOK,
			TrashPath: trashPath,
		}
		
		if err != nil {
			jobItem.Status = core.StatusError
			jobItem.Error = err.Error()
		}
		
		job.Items = append(job.Items, jobItem)
	}
	
	// Save Job
	s.gui.JobMgr.SaveJob(job)
	
	// Remove from shelf
	s.remove(g, v) // Reuse remove logic which handles selection clearing
	
	// Refresh browser if we deleted something in current view
	if s.gui.Browser != nil {
		s.gui.Browser.Refresh()
	}
	
	return nil
}

func (s *Shelf) executePut(g *gocui.Gui, v *gocui.View) error {
	targetDir := s.gui.State.TargetDir
	if targetDir == "" {
		return nil // No target
	}
	
	// Items to put: All items in shelf? Or selected?
	// Spec H3: "p = shelf 선택 항목을 target으로 put"
	// Actually typical lazygit/lazydocker: shelf acts as staging. usually 'p' pastes ALL shelf items?
	// But let's follow the spec: "shelf 선택 항목을" -> "Selection in Shelf".
	// However, if nothing selected, maybe all?
	// Let's assume: if selection exists, put selection. If NO selection, put ALL shelf items.
	
	var targets []store.ShelfItem
	if len(s.selected) > 0 {
		for _, item := range s.gui.State.ShelfItems {
			if _, ok := s.selected[item.ID]; ok {
				targets = append(targets, item)
			}
		}
	} else {
		targets = s.gui.State.ShelfItems
	}
	
	if len(targets) == 0 {
		return nil
	}
	
	job := s.gui.JobMgr.CreateJob(core.JobPut)
	
	for _, item := range targets {
		dst := filepath.Join(targetDir, filepath.Base(item.AbsPath))
		
		// Conflict Check
		finalDst, err := core.ResolveConflict(item.AbsPath, dst, core.PolicyRename) // Default to Rename for now as per MVP safe default or Skip? Spec says "기본 skip" (F2).
		// Wait, Spec says "Basic conflict policy: skip". "Override: skip | rename | overwrite".
		// H3 says "기본 skip로 실행".
		// But "P(대문자) or modal...".
		// Let's implement Skip as default.
		
		finalDst, err = core.ResolveConflict(item.AbsPath, dst, core.PolicySkip)
		if err != nil {
			// Log error
			continue
		}
		
		if finalDst == "" {
			// Skipped
			job.Items = append(job.Items, core.JobItem{
				Src: item.AbsPath,
				Dst: dst,
				Op: string(item.OpMode),
				Status: core.StatusSkipped,
			})
			continue
		}
		
		// Execute Op
		var opErr error
		if item.OpMode == store.OpMove {
			opErr = core.Move(item.AbsPath, finalDst)
		} else {
			if item.Type == "dir" {
				opErr = core.CopyDir(item.AbsPath, finalDst)
			} else {
				opErr = core.CopyFile(item.AbsPath, finalDst)
			}
		}
		
		jobItem := core.JobItem{
			Src: item.AbsPath,
			Dst: finalDst,
			Op: string(item.OpMode),
			CreatedPath: finalDst,
			Status: core.StatusOK,
		}
		
		if opErr != nil {
			jobItem.Status = core.StatusError
			jobItem.Error = opErr.Error()
		}
		
		job.Items = append(job.Items, jobItem)
	}
	
	s.gui.JobMgr.SaveJob(job)
	
	// If Move success, remove from shelf? 
	// Spec doesn't explicitly say to remove from shelf after Put. 
	// But logically Move should remove. Copy should keep?
	// Let's remove successfully moved items.
	
	var newShelf []store.ShelfItem
	for _, item := range s.gui.State.ShelfItems {
		// check if this item was processed successfully as Move
		removed := false
		for _, ji := range job.Items {
			if ji.Src == item.AbsPath && ji.Op == "move" && ji.Status == core.StatusOK {
				removed = true
				break
			}
		}
		if !removed {
			newShelf = append(newShelf, item)
		}
	}
	s.gui.State.ShelfItems = newShelf
	s.selected = make(map[string]struct{}) // Clear selection
	
	s.Update()
	if s.gui.Browser != nil {
		s.gui.Browser.Refresh()
	}
	
	return nil
}
