package storage

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const separator = "---"

var checklistRe = regexp.MustCompile(`(?m)^-\s+\[(x| )\]\s+(.+)$`)

func parseFrontmatter(content string) (Task, error) {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, separator) {
		return Task{}, fmt.Errorf("frontmatter manquant")
	}

	// Retire le premier "---"
	rest := content[len(separator):]
	idx := strings.Index(rest, separator)
	if idx == -1 {
		return Task{}, fmt.Errorf("frontmatter non fermé")
	}

	front := strings.TrimSpace(rest[:idx])
	var task Task
	if err := yaml.Unmarshal([]byte(front), &task); err != nil {
		return Task{}, fmt.Errorf("erreur parsing YAML: %w", err)
	}

	// Parse le corps Markdown (après le second "---")
	body := strings.TrimSpace(rest[idx+len(separator):])
	task.Checklist = parseChecklist(body)

	return task, nil
}

func parseChecklist(body string) []ChecklistItem {
	matches := checklistRe.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return nil
	}
	items := make([]ChecklistItem, 0, len(matches))
	for _, m := range matches {
		items = append(items, ChecklistItem{
			Done: m[1] == "x",
			Text: strings.TrimSpace(m[2]),
		})
	}
	return items
}

// parseFrontmatterSafe est comme parseFrontmatter mais applique des valeurs par
// défaut pour les champs critiques manquants (id, status, title).
func parseFrontmatterSafe(content, filename string) (Task, error) {
	task, err := parseFrontmatter(content)
	if err != nil {
		return Task{}, err
	}
	if task.ID == "" {
		base := filepath.Base(filename)
		task.ID = strings.TrimSuffix(base, filepath.Ext(base))
	}
	if task.Status == "" {
		task.Status = "todo"
	}
	if task.Title == "" {
		task.Title = "Sans titre"
	}
	return task, nil
}

func writeFrontmatter(task Task) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(separator + "\n")

	data, err := yaml.Marshal(task)
	if err != nil {
		return nil, err
	}
	buf.Write(data)
	buf.WriteString(separator + "\n")

	// Écriture de la checklist dans le corps Markdown
	if len(task.Checklist) > 0 {
		buf.WriteString("\n")
		for _, item := range task.Checklist {
			mark := " "
			if item.Done {
				mark = "x"
			}
			fmt.Fprintf(&buf, "- [%s] %s\n", mark, item.Text)
		}
	}

	return buf.Bytes(), nil
}
