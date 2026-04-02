package storage

import "time"

type Task struct {
	ID          string    `yaml:"id"`
	Title       string    `yaml:"title"`
	Status      string    `yaml:"status"`
	Description string    `yaml:"description"`
	Due         string    `yaml:"due,omitempty"`
	Created     time.Time `yaml:"created"`
	SubTasks    []string  `yaml:"sub_tasks,omitempty"`
}

func (t Task) DueDisplay() string {
	if t.Due == "" {
		return ""
	}
	return t.Due
}
