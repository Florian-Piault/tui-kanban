package cli

import (
	"fmt"
	"os"

	"github.com/piflorian/tui-kanban/internal/config"
	"github.com/piflorian/tui-kanban/internal/storage"
)

func runDelete(cfg *config.Config, store *storage.Storage, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: kanban delete <ID>")
		return 1
	}
	id := args[0]
	project := cfg.CurrentProject

	if err := store.DeleteTask(project, id); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur : %v\n", err)
		return 1
	}

	fmt.Printf("Tâche %s supprimée.\n", id)
	return 0
}
