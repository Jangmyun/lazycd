package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ConflictPolicy string

const (
	PolicySkip      ConflictPolicy = "skip"
	PolicyOverwrite ConflictPolicy = "overwrite"
	PolicyRename    ConflictPolicy = "rename"
)

// CheckConflict returns whether the destination exists.
func CheckConflict(dst string) (bool, error) {
	_, err := os.Stat(dst)
	if err == nil {
		return true, nil // Exists
	}
	if os.IsNotExist(err) {
		return false, nil // Does not exist
	}
	return false, err // Other error
}

// ResolveConflict returns the final path to use based on the policy.
// If PolicySkip, returns "" (indicating skip).
// If PolicyRename, returns a new non-conflicting path.
// If PolicyOverwrite, returns original dst (but checks safety).
func ResolveConflict(src, dst string, policy ConflictPolicy) (string, error) {
	exists, err := CheckConflict(dst)
	if err != nil {
		return "", err
	}
	
	if !exists {
		return dst, nil
	}
	
	switch policy {
	case PolicySkip:
		return "", nil // Skip this item
		
	case PolicyOverwrite:
		// Safety check: Cannot overwrite directory with file or vice-versa easily without recursive delete.
		// Spec says: "dst가 폴더면 overwrite 불가"
		dstInfo, err := os.Stat(dst)
		if err != nil {
			return "", err
		}
		if dstInfo.IsDir() {
			return "", fmt.Errorf("cannot overwrite directory '%s'", dst)
		}
		// If src is dir and dst is file -> error? Not explicitly said but usually logic error.
		// For now assume we proceed with overwrite (caller handles backup)
		return dst, nil
		
	case PolicyRename:
		return findFreeName(dst)
		
	default:
		return "", fmt.Errorf("unknown policy: %s", policy)
	}
}

// findFreeName generates name (1).ext, name (2).ext ...
func findFreeName(path string) (string, error) {
	ext := filepath.Ext(path)
	baseNoExt := strings.TrimSuffix(filepath.Base(path), ext)
	dir := filepath.Dir(path)
	
	for i := 1; i < 10000; i++ {
		newName := fmt.Sprintf("%s (%d)%s", baseNoExt, i, ext)
		newPath := filepath.Join(dir, newName)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath, nil
		}
	}
	return "", fmt.Errorf("failed to find free name for %s", path)
}
