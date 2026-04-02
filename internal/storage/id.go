package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func nextID(projectDir string) (string, error) {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "TASK-001", nil
		}
		return "", err
	}

	max := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".md")
		if !strings.HasPrefix(name, "TASK-") {
			continue
		}
		n, err := strconv.Atoi(name[5:])
		if err == nil && n > max {
			max = n
		}
	}

	return fmt.Sprintf("TASK-%03d", max+1), nil
}

func taskFilename(id string) string {
	return id + ".md"
}

func taskPath(projectDir, id string) string {
	return filepath.Join(projectDir, taskFilename(id))
}
