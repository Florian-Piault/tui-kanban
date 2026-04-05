package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/piflorian/tui-kanban/internal/config"
	"github.com/piflorian/tui-kanban/internal/storage"
)

func runAdd(cfg *config.Config, store *storage.Storage, args []string) int {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	typeFlag := fs.String("type", "task", "Type de tâche (task|bug|feat|doc)")
	statusFlag := fs.String("status", "", "Status initial")
	dueFlag := fs.String("due", "", "Date d'échéance (ex: 2026-04-10)")
	projectFlag := fs.String("project", "", "Projet cible")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Usage: kanban add <titre> [--type task|bug|feat|doc] [--status todo] [--due 2026-04-10]")
		return 1
	}
	title := fs.Arg(0)

	if !storage.IsValidType(*typeFlag) {
		fmt.Fprintf(os.Stderr, "Erreur : type %q invalide. Valeurs : task|bug|feat|doc\n", *typeFlag)
		return 1
	}

	status := *statusFlag
	if status == "" {
		if len(cfg.Columns) > 0 {
			status = cfg.Columns[0].ID
		} else {
			status = "todo"
		}
	} else {
		if _, ok := cfg.ColumnByID(status); !ok {
			fmt.Fprintf(os.Stderr, "Erreur : status %q invalide. Valeurs : %s\n",
				status, joinStrings(validStatuses(cfg)))
			return 1
		}
	}

	project := resolveProject(cfg, *projectFlag)

	task := storage.Task{
		Title:  title,
		Type:   storage.NormalizeType(*typeFlag),
		Status: status,
		Due:    *dueFlag,
	}

	created, err := store.SaveTask(project, task)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur interne : %v\n", err)
		return 2
	}

	fmt.Printf("Tâche créée : %s  %q\n", created.ID, created.Title)
	return 0
}
