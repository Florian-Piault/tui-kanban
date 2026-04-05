package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/piflorian/tui-kanban/internal/config"
	"github.com/piflorian/tui-kanban/internal/storage"
)

func runLint(cfg *config.Config, store *storage.Storage, args []string) int {
	fs := flag.NewFlagSet("lint", flag.ContinueOnError)
	project := fs.String("project", "", "Projet cible")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	proj := resolveProject(cfg, *project)
	result := store.LoadAll(proj)

	if len(result.FilesWithErrors) == 0 {
		fmt.Printf("✓ %d tâche(s) chargée(s), aucune erreur\n", len(result.Tasks))
		return 0
	}

	fmt.Fprintf(os.Stderr, "✗ %d fichier(s) avec erreur(s) :\n\n", len(result.FilesWithErrors))
	for _, fe := range result.FilesWithErrors {
		fmt.Fprintf(os.Stderr, "  %s\n    → %s\n\n", fe.Path, fe.Reason)
	}
	return 1
}
