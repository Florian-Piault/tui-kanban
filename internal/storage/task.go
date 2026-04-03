package storage

import (
	"strings"
	"time"
)

const (
	TypeTask = "task"
	TypeBug  = "bug"
	TypeFeat = "feat"
	TypeDoc  = "doc"
)

// AllTypes est la liste ordonnée des types disponibles.
var AllTypes = []string{TypeTask, TypeBug, TypeFeat, TypeDoc}

// typeAliases mappe alias → type normalisé.
var typeAliases = map[string]string{
	"task": TypeTask,
	"bug":  TypeBug,
	"feat": TypeFeat,
	"doc":  TypeDoc,
}

// TypePrefix mappe le type normalisé → préfixe d'ID.
var TypePrefix = map[string]string{
	TypeTask: "TASK",
	TypeBug:  "BUG",
	TypeFeat: "FEAT",
	TypeDoc:  "DOC",
}

// IsValidType retourne true si s est un type ou alias connu.
func IsValidType(s string) bool {
	_, ok := typeAliases[strings.ToLower(s)]
	return ok
}

// NormalizeType convertit un alias en type canonique (défaut : "task").
func NormalizeType(s string) string {
	if t, ok := typeAliases[strings.ToLower(s)]; ok {
		return t
	}
	return TypeTask
}

type ChecklistItem struct {
	Text string
	Done bool
}

type Task struct {
	ID          string    `yaml:"id"`
	Type        string    `yaml:"type,omitempty"`
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
