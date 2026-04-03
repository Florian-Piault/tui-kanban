package storage

import "time"

type ChecklistItem struct {
	Text string
	Done bool
}

type Task struct {
	ID          string    `yaml:"id"`
	Title       string    `yaml:"title"`
	Status      string    `yaml:"status"`
	Description string    `yaml:"description"`
	Due         string    `yaml:"due,omitempty"`
	Created     time.Time `yaml:"created"`
	// Checklist est stocké dans le corps Markdown (pas dans le YAML)
	Checklist []ChecklistItem `yaml:"-"`
}

func (t Task) DueDisplay() string {
	if t.Due == "" {
		return ""
	}
	return t.Due
}

// ChecklistProgress retourne le nombre d'items cochés et le total.
func (t Task) ChecklistProgress() (done, total int) {
	total = len(t.Checklist)
	for _, item := range t.Checklist {
		if item.Done {
			done++
		}
	}
	return
}
