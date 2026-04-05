package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/piflorian/tui-kanban/internal/config"
	"github.com/piflorian/tui-kanban/internal/storage"
)

func runList(cfg *config.Config, store *storage.Storage, args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	statusFlag := fs.String("status", "", "Filtrer par status")
	typeFlag := fs.String("type", "", "Filtrer par type (task|bug|feat|doc)")
	projectFlag := fs.String("project", "", "Projet cible")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *statusFlag != "" {
		if _, ok := cfg.ColumnByID(*statusFlag); !ok {
			fmt.Fprintf(os.Stderr, "Erreur : status %q invalide. Valeurs : %s\n",
				*statusFlag, joinStrings(validStatuses(cfg)))
			return 1
		}
	}

	if *typeFlag != "" && !storage.IsValidType(*typeFlag) {
		fmt.Fprintf(os.Stderr, "Erreur : type %q invalide. Valeurs : task|bug|feat|doc\n", *typeFlag)
		return 1
	}

	project := resolveProject(cfg, *projectFlag)

	var tasks []storage.Task
	if *statusFlag != "" {
		var err error
		tasks, err = store.LoadByStatus(project, *statusFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erreur interne : %v\n", err)
			return 2
		}
	} else {
		tasks = store.LoadAll(project).Tasks
	}

	if *typeFlag != "" {
		normalized := storage.NormalizeType(*typeFlag)
		filtered := tasks[:0]
		for _, t := range tasks {
			if t.Type == normalized {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}

	if len(tasks) == 0 {
		fmt.Fprintln(os.Stderr, "Aucune tâche trouvée.")
		return 0
	}

	printTaskHeader(os.Stdout)
	for _, t := range tasks {
		printTaskRow(os.Stdout, t)
	}
	return 0
}

func joinStrings(ss []string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}
