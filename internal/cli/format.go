package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/piflorian/tui-kanban/internal/storage"
)

const (
	colWidthID     = 10
	colWidthType   = 6
	colWidthStatus = 13
	colWidthDue    = 12
	colWidthTitle  = 50
)

func printTaskHeader(w io.Writer) {
	fmt.Fprintf(w, "%-*s  %-*s  %-*s  %-*s  %s\n",
		colWidthID, "ID",
		colWidthType, "TYPE",
		colWidthStatus, "STATUS",
		colWidthDue, "DUE",
		"TITLE",
	)
	fmt.Fprintf(w, "%s  %s  %s  %s  %s\n",
		strings.Repeat("-", colWidthID),
		strings.Repeat("-", colWidthType),
		strings.Repeat("-", colWidthStatus),
		strings.Repeat("-", colWidthDue),
		strings.Repeat("-", colWidthTitle),
	)
}

func printTaskRow(w io.Writer, t storage.Task) {
	title := t.Title
	if len(title) > colWidthTitle {
		title = title[:colWidthTitle-3] + "..."
	}
	due := t.Due
	fmt.Fprintf(w, "%-*s  %-*s  %-*s  %-*s  %s\n",
		colWidthID, t.ID,
		colWidthType, t.Type,
		colWidthStatus, t.Status,
		colWidthDue, due,
		title,
	)
}

func printTaskDetail(w io.Writer, t storage.Task) {
	fmt.Fprintf(w, "%-12s: %s\n", "ID", t.ID)
	fmt.Fprintf(w, "%-12s: %s\n", "Type", t.Type)
	fmt.Fprintf(w, "%-12s: %s\n", "Titre", t.Title)
	fmt.Fprintf(w, "%-12s: %s\n", "Status", t.Status)
	fmt.Fprintf(w, "%-12s: %s\n", "Créée", t.Created.Format("2006-01-02 15:04"))
	if t.Due != "" {
		fmt.Fprintf(w, "%-12s: %s\n", "Échéance", t.Due)
	}
	if t.Description != "" {
		fmt.Fprintf(w, "\n%-12s:\n", "Description")
		for _, line := range strings.Split(t.Description, "\n") {
			fmt.Fprintf(w, "  %s\n", line)
		}
	}
	if len(t.Checklist) > 0 {
		done, total := t.ChecklistProgress()
		fmt.Fprintf(w, "\n%-12s: %d/%d items complétés\n", "Checklist", done, total)
		for _, item := range t.Checklist {
			check := " "
			if item.Done {
				check = "x"
			}
			fmt.Fprintf(w, "  [%s] %s\n", check, item.Text)
		}
	}
}
