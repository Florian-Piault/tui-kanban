package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Storage struct {
	baseDir string
}

func New(baseDir string) *Storage {
	return &Storage{baseDir: baseDir}
}

func (s *Storage) projectDir(project string) string {
	return filepath.Join(s.baseDir, project)
}

// LoadAll charge toutes les tâches d'un projet, retournées par status.
func (s *Storage) LoadAll(project string) ([]Task, error) {
	dir := s.projectDir(project)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var tasks []Task
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		task, err := parseFrontmatter(string(data))
		if err != nil {
			continue
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// LoadByStatus retourne les tâches d'une colonne/status donné.
func (s *Storage) LoadByStatus(project, status string) ([]Task, error) {
	all, err := s.LoadAll(project)
	if err != nil {
		return nil, err
	}
	var filtered []Task
	for _, t := range all {
		if t.Status == status {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

// SaveTask crée ou met à jour une tâche. Génère l'ID si absent.
func (s *Storage) SaveTask(project string, task Task) (Task, error) {
	dir := s.projectDir(project)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return Task{}, err
	}

	if task.ID == "" {
		id, err := nextID(dir)
		if err != nil {
			return Task{}, fmt.Errorf("génération ID: %w", err)
		}
		task.ID = id
	}
	if task.Created.IsZero() {
		task.Created = time.Now()
	}

	data, err := writeFrontmatter(task)
	if err != nil {
		return Task{}, err
	}

	path := taskPath(dir, task.ID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return Task{}, err
	}
	return task, nil
}

// DeleteTask supprime le fichier d'une tâche.
func (s *Storage) DeleteTask(project, id string) error {
	path := taskPath(s.projectDir(project), id)
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("tâche %s introuvable", id)
	}
	return err
}

// MoveTask change le status d'une tâche (déplacement de colonne).
func (s *Storage) MoveTask(project, id, newStatus string) (Task, error) {
	dir := s.projectDir(project)
	path := taskPath(dir, id)

	data, err := os.ReadFile(path)
	if err != nil {
		return Task{}, fmt.Errorf("tâche %s introuvable: %w", id, err)
	}

	task, err := parseFrontmatter(string(data))
	if err != nil {
		return Task{}, err
	}

	task.Status = newStatus
	return s.SaveTask(project, task)
}

// ListProjects retourne les sous-dossiers du baseDir.
func (s *Storage) ListProjects() ([]string, error) {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var projects []string
	for _, e := range entries {
		if e.IsDir() {
			projects = append(projects, e.Name())
		}
	}
	return projects, nil
}

// GetTask retourne une tâche par ID.
func (s *Storage) GetTask(project, id string) (Task, error) {
	path := taskPath(s.projectDir(project), id)
	data, err := os.ReadFile(path)
	if err != nil {
		return Task{}, fmt.Errorf("tâche %s introuvable: %w", id, err)
	}
	return parseFrontmatter(string(data))
}

// AllTaskIDs retourne tous les IDs de tâches d'un projet.
func (s *Storage) AllTaskIDs(project string) ([]string, error) {
	tasks, err := s.LoadAll(project)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
	}
	return ids, nil
}
