package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type OpMode string

const (
	OpCopy OpMode = "copy"
	OpMove OpMode = "move"
)

type ShelfItem struct {
	ID        string    `json:"id"`
	AbsPath   string    `json:"abs_path"`
	Type      string    `json:"type"` // "file" or "dir"
	OpMode    OpMode    `json:"op_mode"`
	AddedAt   time.Time `json:"added_at"`
	Exists    bool      `json:"exists"` // Runtime flag, might not be needed in JSON but good for cache
}

type State struct {
	LastDir    string      `json:"last_dir"`
	TargetDir  string      `json:"target_dir"`
	ShelfItems []ShelfItem `json:"shelf_items"`
}

func NewState() *State {
	return &State{
		ShelfItems: []ShelfItem{},
	}
}

func (s *State) Save() error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	statePath := filepath.Join(configDir, "state.json")
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0644)
}

func LoadState() (*State, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	statePath := filepath.Join(configDir, "state.json")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return NewState(), nil
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "lazycd")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func NewShelfItem(absPath string, isDir bool) ShelfItem {
	t := "file"
	if isDir {
		t = "dir"
	}
	return ShelfItem{
		ID:      uuid.New().String(),
		AbsPath: absPath,
		Type:    t,
		OpMode:  OpCopy, // Default to copy
		AddedAt: time.Now(),
		Exists:  true,
	}
}
