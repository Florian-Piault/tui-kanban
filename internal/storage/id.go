package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var counterMu sync.Mutex

type stateData struct {
	LastID     int    `json:"last_id"`
	SortMethod string `json:"sort_method,omitempty"`
}

// LoadSortMethod retourne la méthode de tri persistée dans .state.json.
func LoadSortMethod(baseDir string) string {
	statePath := filepath.Join(baseDir, ".state.json")
	var state stateData
	if data, err := os.ReadFile(statePath); err == nil {
		_ = json.Unmarshal(data, &state)
	}
	return state.SortMethod
}

// SaveSortMethod persiste la méthode de tri dans .state.json (écriture atomique).
func SaveSortMethod(baseDir, method string) error {
	counterMu.Lock()
	defer counterMu.Unlock()

	statePath := filepath.Join(baseDir, ".state.json")
	var state stateData
	if data, err := os.ReadFile(statePath); err == nil {
		_ = json.Unmarshal(data, &state)
	}
	state.SortMethod = method

	newData, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	tmpPath := statePath + ".tmp"
	if err := os.WriteFile(tmpPath, newData, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, statePath)
}

// nextID génère un nouvel ID unique en incrémentant le compteur global dans .state.json.
// Écriture atomique via fichier temporaire + rename pour éviter la corruption.
// counterMu protège contre les accès concurrents dans le même processus.
func nextID(baseDir, taskType string) (string, error) {
	counterMu.Lock()
	defer counterMu.Unlock()

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", err
	}

	statePath := filepath.Join(baseDir, ".state.json")

	var state stateData
	if data, err := os.ReadFile(statePath); err == nil {
		_ = json.Unmarshal(data, &state)
	}

	state.LastID++

	newData, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return "", err
	}

	// Écriture atomique : temp + rename
	tmpPath := statePath + ".tmp"
	if err := os.WriteFile(tmpPath, newData, 0644); err != nil {
		return "", err
	}
	if err := os.Rename(tmpPath, statePath); err != nil {
		return "", err
	}

	prefix := TypePrefix[NormalizeType(taskType)]
	return fmt.Sprintf("%s-%03d", prefix, state.LastID), nil
}

func taskFilename(id string) string {
	return id + ".md"
}

func taskPath(projectDir, id string) string {
	return filepath.Join(projectDir, taskFilename(id))
}
