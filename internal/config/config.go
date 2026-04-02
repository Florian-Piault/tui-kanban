package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Column struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

type Config struct {
	Columns        []Column `yaml:"columns"`
	ProjectsDir    string   `yaml:"projects_dir"`
	CurrentProject string   `yaml:"current_project"`
}

func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kanban", "config.yaml")
}

func DefaultProjectsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kanban")
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := Default()
			if saveErr := Save(cfg, path); saveErr != nil {
				return nil, saveErr
			}
			return cfg, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if len(cfg.Columns) == 0 {
		cfg.Columns = defaultColumns()
	}
	if cfg.ProjectsDir == "" {
		cfg.ProjectsDir = DefaultProjectsDir()
	}
	if cfg.CurrentProject == "" {
		cfg.CurrentProject = "default"
	}
	return &cfg, nil
}

func Save(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func Default() *Config {
	return &Config{
		Columns:        defaultColumns(),
		ProjectsDir:    DefaultProjectsDir(),
		CurrentProject: "default",
	}
}

func defaultColumns() []Column {
	return []Column{
		{ID: "todo", Name: "À faire"},
		{ID: "in-progress", Name: "En cours"},
		{ID: "done", Name: "Terminé"},
	}
}

func (c *Config) ColumnByID(id string) (Column, bool) {
	for _, col := range c.Columns {
		if col.ID == id {
			return col, true
		}
	}
	return Column{}, false
}

func (c *Config) ProjectDir() string {
	dir := c.ProjectsDir
	if dir == "" {
		dir = DefaultProjectsDir()
	}
	return filepath.Join(dir, c.CurrentProject)
}
