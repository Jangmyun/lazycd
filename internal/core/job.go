package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/uuid"
)

type JobType string

const (
	JobPut    JobType = "put"
	JobDelete JobType = "delete"
)

type JobItemStatus string

const (
	StatusOK      JobItemStatus = "ok"
	StatusSkipped JobItemStatus = "skipped"
	StatusError   JobItemStatus = "error"
)

type JobItem struct {
	Src         string        `json:"src"`
	Dst         string        `json:"dst,omitempty"`
	Op          string        `json:"op"` // copy, move, delete
	Status      JobItemStatus `json:"status"`
	Error       string        `json:"error,omitempty"`
	CreatedPath string        `json:"created_path,omitempty"` // For copy/move
	BackupPath  string        `json:"backup_path,omitempty"`  // For overwrite
	TrashPath   string        `json:"trash_path,omitempty"`   // For delete
}

type Job struct {
	ID        string    `json:"id"`
	Type      JobType   `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Items     []JobItem `json:"items"`
}

type JobManager struct {
	JobsDir string
}

func NewJobManager() (*JobManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	jobsDir := filepath.Join(home, ".config", "lazycd", "jobs")
	if err := os.MkdirAll(jobsDir, 0755); err != nil {
		return nil, err
	}
	return &JobManager{JobsDir: jobsDir}, nil
}

func (jm *JobManager) CreateJob(jType JobType) *Job {
	return &Job{
		ID:        uuid.New().String(),
		Type:      jType,
		CreatedAt: time.Now(),
		Items:     []JobItem{},
	}
}

func (jm *JobManager) SaveJob(job *Job) error {
	path := filepath.Join(jm.JobsDir, job.ID+".json")
	data, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (jm *JobManager) GetRecentJobs(n int) ([]*Job, error) {
	entries, err := os.ReadDir(jm.JobsDir)
	if err != nil {
		return nil, err
	}

	var jobs []*Job
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".json" {
			data, err := os.ReadFile(filepath.Join(jm.JobsDir, entry.Name()))
			if err != nil {
				continue
			}
			var job Job
			if err := json.Unmarshal(data, &job); err == nil {
				jobs = append(jobs, &job)
			}
		}
	}

	// Sort by CreatedAt desc
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.After(jobs[j].CreatedAt)
	})

	if len(jobs) > n {
		jobs = jobs[:n]
	}
	return jobs, nil
}

// Undo reverses the operations in the job.
func (jm *JobManager) Undo(job *Job) error {
	// Reverse items to undo in reverse order
	for i := len(job.Items) - 1; i >= 0; i-- {
		item := job.Items[i]
		if item.Status != StatusOK {
			continue
		}

		var err error
		switch item.Op {
		case "copy":
			// Delete created file/dir
			if item.CreatedPath != "" {
				err = os.RemoveAll(item.CreatedPath)
			}
		case "move":
			// Move back from dst check (CreatedPath) to src
			if item.CreatedPath != "" {
				// Try move back
				// If src exists (unexpected?), we might need conflict resolution logic even for undo?
				// For MVP, simplistic fallback:
				err = Move(item.CreatedPath, item.Src)
			}
		case "delete":
			// Move from trash back to src
			if item.TrashPath != "" {
				err = Move(item.TrashPath, item.Src)
			}
		}
		
		// Overwrite undo
		if item.BackupPath != "" {
			// Restore backup to dst (overwrite current dst)
			// Move backup -> dst
			// We might need to delete current dst first if it's there
			if item.Dst != "" {
				_ = os.RemoveAll(item.Dst) // Remove the "new" file
				err = Move(item.BackupPath, item.Dst)
			}
		}

		if err != nil {
			// Log error but continue?
			fmt.Printf("Undo error for %s: %v\n", item.Src, err)
		}
	}
	
	// Remove job file after undo? Or mark as undone?
	// For now, let's remove it to prevent double undo
	path := filepath.Join(jm.JobsDir, job.ID+".json")
	return os.Remove(path)
}
