package command

// ArgKind définit le type d'un argument pour l'autocomplétion.
type ArgKind int

const (
	ArgFree        ArgKind = iota // texte libre
	ArgTaskID                     // complété avec les IDs existants
	ArgColumnID                   // complété avec les IDs de colonnes
	ArgProjectName                // complété avec les projets disponibles
)

// ArgSpec décrit un argument d'une commande.
type ArgSpec struct {
	Name     string
	Kind     ArgKind
	Required bool
}

// CommandDef décrit une commande slash.
type CommandDef struct {
	Name    string
	Aliases []string
	Args    []ArgSpec
	Help    string
}

// Registry liste toutes les commandes disponibles.
var Registry = []CommandDef{
	{
		Name: "add",
		Args: []ArgSpec{
			{Name: "titre", Kind: ArgFree, Required: true},
		},
		Help: "Ajouter une tâche dans la colonne active  (-q : création rapide sans modal)",
	},
	{
		Name: "edit",
		Args: []ArgSpec{
			{Name: "id", Kind: ArgTaskID, Required: true},
		},
		Help: "Éditer une tâche existante",
	},
	{
		Name:    "delete",
		Aliases: []string{"rm"},
		Args: []ArgSpec{
			{Name: "id", Kind: ArgTaskID, Required: true},
		},
		Help: "Supprimer une tâche",
	},
	{
		Name: "move",
		Args: []ArgSpec{
			{Name: "id", Kind: ArgTaskID, Required: true},
			{Name: "colonne", Kind: ArgColumnID, Required: true},
		},
		Help: "Déplacer une tâche vers une colonne",
	},
	{
		Name: "sub-add",
		Args: []ArgSpec{
			{Name: "texte", Kind: ArgFree, Required: true},
		},
		Help: "Ajouter une sous-tâche à la tâche sélectionnée",
	},
	{
		Name: "project",
		Args: []ArgSpec{
			{Name: "nom", Kind: ArgProjectName, Required: true},
		},
		Help: "Changer de projet actif",
	},
	{
		Name: "column-add",
		Args: []ArgSpec{
			{Name: "nom", Kind: ArgFree, Required: true},
		},
		Help: "Ajouter une nouvelle colonne (ex: /column-add Revue)",
	},
	{
		Name:    "column-rename",
		Aliases: []string{"column-mv"},
		Args: []ArgSpec{
			{Name: "id", Kind: ArgColumnID, Required: true},
			{Name: "nouveau-nom", Kind: ArgFree, Required: true},
		},
		Help: "Renommer une colonne (ex: /column-rename todo À traiter)",
	},
	{
		Name:    "column-delete",
		Aliases: []string{"column-rm"},
		Args: []ArgSpec{
			{Name: "id", Kind: ArgColumnID, Required: true},
		},
		Help: "Supprimer une colonne",
	},
	{
		Name: "column-left",
		Args: []ArgSpec{
			{Name: "id", Kind: ArgColumnID, Required: true},
		},
		Help: "Déplacer une colonne vers la gauche",
	},
	{
		Name: "column-right",
		Args: []ArgSpec{
			{Name: "id", Kind: ArgColumnID, Required: true},
		},
		Help: "Déplacer une colonne vers la droite",
	},
	{
		Name: "projects-dir",
		Args: []ArgSpec{
			{Name: "chemin", Kind: ArgFree, Required: true},
		},
		Help: "Changer le répertoire de base des projets (ex: /projects-dir ~/work/kanban)",
	},
	{
		Name:    "quit",
		Aliases: []string{"q"},
		Help:    "Quitter l'application",
	},
	{
		Name: "help",
		Help: "Afficher l'aide",
	},
}

// Resolve retourne la CommandDef correspondant à un nom ou alias.
func Resolve(name string) (CommandDef, bool) {
	for _, cmd := range Registry {
		if cmd.Name == name {
			return cmd, true
		}
		for _, alias := range cmd.Aliases {
			if alias == name {
				return cmd, true
			}
		}
	}
	return CommandDef{}, false
}

// AllNames retourne tous les noms + aliases de commandes.
func AllNames() []string {
	var names []string
	for _, cmd := range Registry {
		names = append(names, cmd.Name)
	}
	return names
}
