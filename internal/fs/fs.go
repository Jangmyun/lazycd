package fs

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
)

// FileItem represents a file or directory in the file system
type FileItem struct {
	Name      string
	Path      string
	IsDir     bool
	Size      int64
	Mode      os.FileMode
}

// ListDir returns a sorted list of files in the directory.
// Directories come first, then files. Both are sorted alphabetically.
func ListDir(path string) ([]FileItem, error) {
	resolvedPath, err := ResolvePath(path)
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(resolvedPath)
	if err != nil {
		return nil, err
	}

	var items []FileItem
	for _, f := range files {
		items = append(items, FileItem{
			Name:  f.Name(),
			Path:  filepath.Join(resolvedPath, f.Name()),
			IsDir: f.IsDir(),
			Size:  f.Size(),
			Mode:  f.Mode(),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].IsDir != items[j].IsDir {
			return items[i].IsDir // Directories first
		}
		return items[i].Name < items[j].Name
	})

	return items, nil
}

// ResolvePath expands ~ to the user's home directory and returns the absolute path.
func ResolvePath(path string) (string, error) {
	if path == "~" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		return usr.HomeDir, nil
	} else if strings.HasPrefix(path, "~/") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		return filepath.Join(usr.HomeDir, path[2:]), nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return absPath, nil
}
