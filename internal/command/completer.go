package command

import (
	"strings"
)

// Context fourni au completer pour les suggestions contextuelles.
type CompletionContext struct {
	TaskIDs    []string
	TaskTitles map[string]string
	ColumnIDs  []string
	ProjectIDs []string
}

// Suggestion représente une entrée d'autocomplétion avec une valeur (appliquée) et un label (affiché).
type Suggestion struct {
	Value string
	Label string
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
		return toSuggestions(commandNames(""))
	}

	// Un seul token et pas d'espace trailing → compléter le nom de commande
	if len(tokens) == 1 && !trailingSpace {
		return toSuggestions(commandNames(tokens[0]))
	}

	// Commande connue, compléter les args
	cmdName := tokens[0]
	def, ok := Resolve(cmdName)
	if !ok {
		return toSuggestions(commandNames(cmdName))
	}

	// Déterminer l'index de l'arg courant
	argIdx := len(tokens) - 1
	if trailingSpace {
		argIdx = len(tokens) - 1
	} else {
		argIdx = len(tokens) - 2
	}

	if argIdx < 0 || argIdx >= len(def.Args) {
		return nil
	}

	spec := def.Args[argIdx]
	var prefix string
	if !trailingSpace && len(tokens) > 1 {
		prefix = tokens[len(tokens)-1]
	}

	switch spec.Kind {
	case ArgTaskID:
		return taskSuggestions(filterPrefix(ctx.TaskIDs, prefix), ctx.TaskTitles)
	case ArgColumnID:
		return toSuggestions(filterPrefix(ctx.ColumnIDs, prefix))
	case ArgProjectName:
		return toSuggestions(filterPrefix(ctx.ProjectIDs, prefix))
	case ArgFree:
		return nil
	}
	return nil
}

func taskSuggestions(ids []string, titles map[string]string) []Suggestion {
	result := make([]Suggestion, len(ids))
	for i, id := range ids {
		label := id
		if t, ok := titles[id]; ok && t != "" {
			full := id + ": " + t
			if len(full) > 35 {
				full = full[:32] + "…"
			}
			label = full
		}
		result[i] = Suggestion{Value: id, Label: label}
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

func commandNames(prefix string) []string {
	var result []string
	for _, cmd := range Registry {
		if strings.HasPrefix(cmd.Name, prefix) {
			result = append(result, cmd.Name)
		}
	}
	return result
}

func filterPrefix(list []string, prefix string) []string {
	if prefix == "" {
		return list
	}
	var result []string
	for _, s := range list {
		if strings.HasPrefix(s, prefix) {
			result = append(result, s)
		}
	}
	return result
}
