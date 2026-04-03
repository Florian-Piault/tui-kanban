package command

import (
	"strings"

	"github.com/sahilm/fuzzy"
)

// CompletionContext fourni au completer pour les suggestions contextuelles.
type CompletionContext struct {
	TaskIDs    []string
	TaskTitles map[string]string
	ColumnIDs  []string
	ProjectIDs []string
}

// Suggestion représente une entrée d'autocomplétion avec une valeur (appliquée) et un label (affiché).
type Suggestion struct {
	Value          string
	Label          string
	MatchedIndexes []int // positions dans Label qui ont matché
}

// Complete retourne les suggestions pour l'input courant.
func Complete(input string, ctx CompletionContext) []Suggestion {
	if !strings.HasPrefix(input, "/") {
		return nil
	}

	tokens := Tokenize(input)
	trailingSpace := len(input) > 1 && input[len(input)-1] == ' '

	// Aucun token après "/" → compléter le nom de commande
	if len(tokens) == 0 {
		return commandSuggestions("")
	}

	// Un seul token et pas d'espace trailing → compléter le nom de commande
	if len(tokens) == 1 && !trailingSpace {
		return commandSuggestions(tokens[0])
	}

	// Commande connue, compléter les args
	cmdName := tokens[0]
	def, ok := Resolve(cmdName)
	if !ok {
		return commandSuggestions(cmdName)
	}

	// Déterminer l'index de l'arg courant
	argIdx := len(tokens) - 1
	if !trailingSpace {
		argIdx = len(tokens) - 2
	}

	if argIdx < 0 || argIdx >= len(def.Args) {
		return nil
	}

	spec := def.Args[argIdx]
	var pattern string
	if !trailingSpace && len(tokens) > 1 {
		pattern = tokens[len(tokens)-1]
	}

	switch spec.Kind {
	case ArgTaskID:
		return taskSuggestions(ctx, pattern)
	case ArgColumnID:
		return fuzzyFilter(ctx.ColumnIDs, pattern)
	case ArgProjectName:
		return fuzzyFilter(ctx.ProjectIDs, pattern)
	case ArgFree:
		return nil
	}
	return nil
}

func commandSuggestions(pattern string) []Suggestion {
	names := make([]string, len(Registry))
	for i, cmd := range Registry {
		names[i] = cmd.Name
	}
	if pattern == "" {
		return toSuggestions(names)
	}
	matches := fuzzy.Find(pattern, names)
	result := make([]Suggestion, len(matches))
	for i, m := range matches {
		result[i] = Suggestion{
			Value:          m.Str,
			Label:          m.Str,
			MatchedIndexes: m.MatchedIndexes,
		}
	}
	return result
}

func taskSuggestions(ctx CompletionContext, pattern string) []Suggestion {
	labels := make([]string, len(ctx.TaskIDs))
	for i, id := range ctx.TaskIDs {
		label := id
		if t, ok := ctx.TaskTitles[id]; ok && t != "" {
			full := id + ": " + t
			runes := []rune(full)
			if len(runes) > 35 {
				full = string(runes[:32]) + "…"
			}
			label = full
		}
		labels[i] = label
	}

	if pattern == "" {
		result := make([]Suggestion, len(ctx.TaskIDs))
		for i, id := range ctx.TaskIDs {
			result[i] = Suggestion{Value: id, Label: labels[i]}
		}
		return result
	}

	matches := fuzzy.Find(pattern, labels)
	result := make([]Suggestion, len(matches))
	for i, m := range matches {
		result[i] = Suggestion{
			Value:          ctx.TaskIDs[m.Index],
			Label:          m.Str,
			MatchedIndexes: m.MatchedIndexes,
		}
	}
	return result
}

func fuzzyFilter(list []string, pattern string) []Suggestion {
	if pattern == "" {
		return toSuggestions(list)
	}
	matches := fuzzy.Find(pattern, list)
	result := make([]Suggestion, len(matches))
	for i, m := range matches {
		result[i] = Suggestion{
			Value:          m.Str,
			Label:          m.Str,
			MatchedIndexes: m.MatchedIndexes,
		}
	}
	return result
}

func toSuggestions(values []string) []Suggestion {
	result := make([]Suggestion, len(values))
	for i, v := range values {
		result[i] = Suggestion{Value: v, Label: v}
	}
	return result
}
