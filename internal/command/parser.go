package command

import (
	"fmt"
	"strings"
)

// ParsedCommand est le résultat du parsing d'une entrée slash.
type ParsedCommand struct {
	Name  string
	Def   CommandDef
	Args  []string
	Flags map[string]bool
}

// Parse tokenise et résout une entrée comme "/add -q Mon titre".
func Parse(input string) (ParsedCommand, error) {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "/") {
		return ParsedCommand{}, fmt.Errorf("les commandes doivent commencer par /")
	}

	input = input[1:] // retire le "/"
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return ParsedCommand{}, fmt.Errorf("commande vide")
	}

	name := parts[0]

	// Séparer flags et args positionnels
	flags := map[string]bool{}
	var cleanArgs []string
	for _, p := range parts[1:] {
		if strings.HasPrefix(p, "-") {
			flags[strings.TrimPrefix(p, "-")] = true
		} else {
			cleanArgs = append(cleanArgs, p)
		}
	}

	def, ok := Resolve(name)
	if !ok {
		return ParsedCommand{}, fmt.Errorf("commande inconnue : %q", name)
	}

	// Si le dernier arg est ArgFree, regrouper tous les tokens restants en une seule chaîne.
	// Ex: "/add Mon titre"              → ["Mon titre"]
	// Ex: "/column-rename todo À faire" → ["todo", "À faire"]
	if len(def.Args) > 0 && def.Args[len(def.Args)-1].Kind == ArgFree && len(cleanArgs) >= len(def.Args) {
		joined := strings.Join(cleanArgs[len(def.Args)-1:], " ")
		cleanArgs = append(cleanArgs[:len(def.Args)-1], joined)
	}

	// Vérifier les args requis (sauf si un flag dispense de l'arg, ex: -q sans titre)
	for i, spec := range def.Args {
		if spec.Required && i >= len(cleanArgs) && !flags["q"] {
			return ParsedCommand{}, fmt.Errorf("argument requis manquant : <%s>", spec.Name)
		}
	}

	return ParsedCommand{
		Name:  def.Name,
		Def:   def,
		Args:  cleanArgs,
		Flags: flags,
	}, nil
}

// Tokenize découpe l'input en tokens sans valider (ignore les flags).
func Tokenize(input string) []string {
	if strings.HasPrefix(input, "/") {
		input = input[1:]
	}
	var result []string
	for _, f := range strings.Fields(input) {
		if !strings.HasPrefix(f, "-") {
			result = append(result, f)
		}
	}
	return result
}
