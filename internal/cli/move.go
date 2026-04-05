package cli

import (
	"fmt"
	"os"

	"github.com/piflorian/tui-kanban/internal/config"
	"github.com/piflorian/tui-kanban/internal/storage"
)

func runMove(cfg *config.Config, store *storage.Storage, args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: kanban move <ID> <status>")
		return 1
	}
	id := args[0]
	newStatus := args[1]

	if _, ok := cfg.ColumnByID(newStatus); !ok {
		fmt.Fprintf(os.Stderr, "Erreur : status %q invalide. Valeurs : %s\n",
			newStatus, joinStrings(validStatuses(cfg)))
		return 1
	}

	project := cfg.CurrentProject
	_, err := store.MoveTask(project, id, newStatus)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur : %v\n", err)
		return 1
	}

	fmt.Printf("%s déplacé vers %q\n", id, newStatus)
	return 0
}

func runDone(cfg *config.Config, store *storage.Storage, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: kanban done <ID>")
		return 1
	}
	return runMove(cfg, store, []string{args[0], doneStatus(cfg)})
}
