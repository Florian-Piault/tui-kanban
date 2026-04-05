package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/piflorian/tui-kanban/internal/config"
	"github.com/piflorian/tui-kanban/internal/storage"
)

// Run est le point d'entrée unique du mode CLI.
// args = os.Args[1:]
// Retourne : 0 succès, 1 erreur utilisateur, 2 erreur interne.
func Run(cfg *config.Config, store *storage.Storage, args []string) int {
	if len(args) == 0 {
		printHelp(cfg)
		return 0
	}
	switch args[0] {
	case "add":
		return runAdd(cfg, store, args[1:])
	case "list", "ls":
		return runList(cfg, store, args[1:])
	case "move":
		return runMove(cfg, store, args[1:])
	case "done":
		return runDone(cfg, store, args[1:])
	case "show":
		return runShow(cfg, store, args[1:])
	case "delete", "rm":
		return runDelete(cfg, store, args[1:])
	case "lint":
		return runLint(cfg, store, args[1:])
	case "help", "--help", "-h":
		printHelp(cfg)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Commande inconnue : %q\n\n", args[0])
		printHelp(cfg)
		return 1
	}
}

func printHelp(cfg *config.Config) {
	statuses := make([]string, len(cfg.Columns))
	for i, col := range cfg.Columns {
		statuses[i] = col.ID
	}
	statusList := strings.Join(statuses, "|")

	fmt.Printf(`Usage: kanban <commande> [options]

Commandes:
  add <titre>    Créer une nouvelle tâche
  list           Lister les tâches
  move <ID> <status>  Déplacer une tâche
  done <ID>      Marquer une tâche comme terminée
  show <ID>      Afficher les détails d'une tâche
  delete <ID>    Supprimer une tâche
  lint           Vérifier l'intégrité des fichiers de tâches
  help           Afficher cette aide

Options communes:
  --project      Projet cible (défaut: %s)

Statuts disponibles: %s
Types disponibles: task|bug|feat|doc
`, cfg.CurrentProject, statusList)
}

// resolveProject retourne le projet actif.
func resolveProject(cfg *config.Config, flagVal string) string {
	if flagVal != "" {
		return flagVal
	}
	return cfg.CurrentProject
}

// doneStatus retourne l'ID du statut "terminé" (cherche "done", sinon dernier).
func doneStatus(cfg *config.Config) string {
	for _, col := range cfg.Columns {
		if col.ID == "done" {
			return col.ID
		}
	}
	if len(cfg.Columns) > 0 {
		return cfg.Columns[len(cfg.Columns)-1].ID
	}
	return "done"
}

// validStatuses retourne la liste des IDs de colonnes valides.
func validStatuses(cfg *config.Config) []string {
	ids := make([]string, len(cfg.Columns))
	for i, col := range cfg.Columns {
		ids[i] = col.ID
	}
	return ids
}
