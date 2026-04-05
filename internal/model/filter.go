package model

import (
	"sort"
	"strconv"
	"strings"

	"github.com/piflorian/tui-kanban/internal/storage"
)

// SortType définit les modes de tri disponibles.
type SortType string

const (
	SortDate     SortType = "date"
	SortPriority SortType = "priority"
	SortTitle    SortType = "title"
)

var sortCycle = []SortType{SortDate, SortPriority, SortTitle}

// FilterQuery représente une requête de filtrage cumulatif.
type FilterQuery struct {
	Text     string   // recherche textuelle libre
	Type     string   // is:bug, is:feat, etc.
	Priority int      // pri:1, pri:2, pri:3
	Tags     []string // #tag
}

func (f FilterQuery) IsEmpty() bool {
	return f.Text == "" && f.Type == "" && f.Priority == 0 && len(f.Tags) == 0
}

// String retourne une représentation lisible de la requête (pour le badge UI).
func (f FilterQuery) String() string {
	var parts []string
	if f.Type != "" {
		parts = append(parts, "is:"+f.Type)
	}
	if f.Priority != 0 {
		parts = append(parts, "pri:"+strconv.Itoa(f.Priority))
	}
	for _, t := range f.Tags {
		parts = append(parts, "#"+t)
	}
	if f.Text != "" {
		parts = append(parts, f.Text)
	}
	return strings.Join(parts, " ")
}

// ParseSearchInput parse une chaîne de recherche avec préfixes optionnels.
// Exemples : "is:bug", "pri:2 fix crash", "is:feat #backend auth"
func ParseSearchInput(input string) FilterQuery {
	words := strings.Fields(input)
	q := FilterQuery{}
	var freeText []string
	for _, word := range words {
		switch {
		case strings.HasPrefix(word, "is:"):
			q.Type = strings.TrimPrefix(word, "is:")
		case strings.HasPrefix(word, "pri:"):
			p, _ := strconv.Atoi(strings.TrimPrefix(word, "pri:"))
			q.Priority = p
		case strings.HasPrefix(word, "#"):
			q.Tags = append(q.Tags, strings.TrimPrefix(word, "#"))
		default:
			freeText = append(freeText, word)
		}
	}
	q.Text = strings.Join(freeText, " ")
	return q
}

func matchesFilter(t storage.Task, q FilterQuery) bool {
	if q.Type != "" && !strings.EqualFold(t.Type, q.Type) {
		return false
	}
	if q.Priority != 0 && t.Priority != q.Priority {
		return false
	}
	if q.Text != "" {
		lower := strings.ToLower(q.Text)
		if !strings.Contains(strings.ToLower(t.Title), lower) &&
			!strings.Contains(strings.ToLower(t.Description), lower) &&
			!strings.Contains(strings.ToLower(t.ID), lower) {
			return false
		}
	}
	return true
}

// ApplyFilter retourne les tâches correspondant à la requête.
// Retourne la slice d'origine si la requête est vide.
func ApplyFilter(tasks []storage.Task, q FilterQuery) []storage.Task {
	if q.IsEmpty() {
		return tasks
	}
	var result []storage.Task
	for _, t := range tasks {
		if matchesFilter(t, q) {
			result = append(result, t)
		}
	}
	return result
}

// SortTasks retourne une nouvelle slice triée selon la méthode choisie.
func SortTasks(tasks []storage.Task, s SortType) []storage.Task {
	result := make([]storage.Task, len(tasks))
	copy(result, tasks)
	switch s {
	case SortPriority:
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].Priority > result[j].Priority
		})
	case SortTitle:
		sort.SliceStable(result, func(i, j int) bool {
			return strings.ToLower(result[i].Title) < strings.ToLower(result[j].Title)
		})
	default: // SortDate
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].Created.Before(result[j].Created)
		})
	}
	return result
}

// NextSort retourne le prochain mode de tri dans le cycle.
func NextSort(current SortType) SortType {
	for i, s := range sortCycle {
		if s == current {
			return sortCycle[(i+1)%len(sortCycle)]
		}
	}
	return SortDate
}

// PrepareView filtre puis trie les tâches avant rendu.
func PrepareView(tasks []storage.Task, q FilterQuery, s SortType) []storage.Task {
	return SortTasks(ApplyFilter(tasks, q), s)
}
