package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// CopyFile copies a file from src to dst, preserving attributes if possible.
func CopyFile(src, dst string) error {
	sourceFileStat, err := os.Lstat(src)
	if err != nil {
		return err
	}

	// Handle symlinks
	if sourceFileStat.Mode()&os.ModeSymlink != 0 {
		linkTarget, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(linkTarget, dst)
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// Create destination
	destination, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sourceFileStat.Mode())
	if err != nil {
		return err
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return err
	}

	// Preserve timestamps
	return os.Chtimes(dst, time.Now(), sourceFileStat.ModTime())
}

// CopyDir recursively copies a directory tree.
func CopyDir(src, dst string) error {
	srcStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !srcStat.IsDir() {
		return fmt.Errorf("source %s is not a directory", src)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcStat.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// Move tries to rename, falls back to copy+delete.
func Move(src, dst string) error {
	// Try atomic rename
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}
	
	// Fallback to Copy + Delete
	// Check if source is dir
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	if info.IsDir() {
		if err := CopyDir(src, dst); err != nil {
			return err // If copy fails, do not delete
		}
	} else {
		if err := CopyFile(src, dst); err != nil {
			return err
		}
	}
	
	// Delete source
	return os.RemoveAll(src)
}

// GetTrashPath determines where to move deleted items: ~/.config/lazycd/trash/<jobID>/<filename>
func GetTrashPath(jobID, originalPath string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	base := filepath.Base(originalPath)
	return filepath.Join(home, ".config", "lazycd", "trash", jobID, base), nil
}

// DeleteToTrash moves the item to the trash location for the given job.
func DeleteToTrash(src, jobID string) (string, error) {
	trashPath, err := GetTrashPath(jobID, src)
	if err != nil {
		return "", err
	}
	
	// Ensure trash dir exists
	trashDir := filepath.Dir(trashPath)
	if err := os.MkdirAll(trashDir, 0755); err != nil {
		return "", err
	}
	
	// Use Move logic
	if err := Move(src, trashPath); err != nil {
		return "", err
	}
	
	return trashPath, nil
}
