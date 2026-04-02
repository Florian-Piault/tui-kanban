package command

import (
	"fmt"
	"strings"
)

// ParsedCommand est le résultat du parsing d'une entrée slash.
type ParsedCommand struct {
	Name string
	Def  CommandDef
	Args []string
}

// Parse tokenise et résout une entrée comme "/add Mon titre".
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
	args := parts[1:]

	def, ok := Resolve(name)
	if !ok {
		return ParsedCommand{}, fmt.Errorf("commande inconnue : %q", name)
	}

	// Pour les commandes à arg texte libre, regrouper les parties restantes
	if len(def.Args) > 0 && len(def.Args) == 1 && def.Args[0].Kind == ArgFree && len(args) > 0 {
		args = []string{strings.Join(args, " ")}
	}

	// Vérifier les args requis
	for i, spec := range def.Args {
		if spec.Required && i >= len(args) {
			return ParsedCommand{}, fmt.Errorf("argument requis manquant : <%s>", spec.Name)
		}
	}

	return ParsedCommand{
		Name: def.Name,
		Def:  def,
		Args: args,
	}, nil
}

// Tokenize découpe l'input en tokens sans valider.
func Tokenize(input string) []string {
	if strings.HasPrefix(input, "/") {
		input = input[1:]
	}
	return strings.Fields(input)
}
