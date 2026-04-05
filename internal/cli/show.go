package cli

import (
	"fmt"
	"os"

	"github.com/piflorian/tui-kanban/internal/config"
	"github.com/piflorian/tui-kanban/internal/storage"
)

func runShow(cfg *config.Config, store *storage.Storage, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: kanban show <ID>")
		return 1
	}
	id := args[0]
	project := cfg.CurrentProject

	task, err := store.GetTask(project, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur : %v\n", err)
		return 1
	}

	printTaskDetail(os.Stdout, task)
	return 0
}
