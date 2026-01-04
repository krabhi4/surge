package downloader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DownloadState represents persisted download state for resume
type DownloadState struct {
	ID         int    `json:"id"`
	URL        string `json:"url"`
	DestPath   string `json:"dest_path"`
	TotalSize  int64  `json:"total_size"`
	Downloaded int64  `json:"downloaded"`
	Tasks      []Task `json:"tasks"` // Remaining tasks
	Filename   string `json:"filename"`
}

// getStatePath returns the path to the state file
func getStatePath(destPath string, id int) string {
	dir := filepath.Dir(destPath)
	return filepath.Join(dir, ".surge", fmt.Sprintf("%d.json", id))
}

// SaveState saves download state to .surge/ID.json
func SaveState(destPath string, state *DownloadState) error {
	statePath := getStatePath(destPath, state.ID)

	// Create .surge directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(statePath), 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// LoadState loads download state from .surge/ID.json
func LoadState(destPath string, id int) (*DownloadState, error) {
	statePath := getStatePath(destPath, id)

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state DownloadState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// DeleteState removes the state file after successful completion
func DeleteState(destPath string, id int) error {
	statePath := getStatePath(destPath, id)

	if err := os.Remove(statePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete state file: %w", err)
	}

	// Try to remove the .surge directory if it's empty
	_ = os.Remove(filepath.Dir(statePath))

	return nil
}
