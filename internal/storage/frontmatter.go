package storage

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const separator = "---"

func parseFrontmatter(content string) (Task, error) {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, separator) {
		return Task{}, fmt.Errorf("frontmatter manquant")
	}

	// Retire le premier "---"
	content = content[len(separator):]
	idx := strings.Index(content, separator)
	if idx == -1 {
		return Task{}, fmt.Errorf("frontmatter non fermé")
	}

	front := strings.TrimSpace(content[:idx])
	var task Task
	if err := yaml.Unmarshal([]byte(front), &task); err != nil {
		return Task{}, fmt.Errorf("erreur parsing YAML: %w", err)
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
	return buf.Bytes(), nil
}
