package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

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

// slugify transforme un nom en identifiant kebab-case ASCII.
func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case unicode.IsSpace(r) || r == '_':
			b.WriteRune('-')
		case r == '-':
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

// AddColumn ajoute une nouvelle colonne. Retourne une erreur si l'ID existe déjà.
func (c *Config) AddColumn(name string) (Column, error) {
	id := slugify(name)
	if id == "" {
		return Column{}, fmt.Errorf("nom de colonne invalide")
	}
	for _, col := range c.Columns {
		if col.ID == id {
			return Column{}, fmt.Errorf("une colonne avec l'ID %q existe déjà", id)
		}
	}
	col := Column{ID: id, Name: name}
	c.Columns = append(c.Columns, col)
	return col, nil
}

// RenameColumn renomme une colonne existante par son ID.
func (c *Config) RenameColumn(id, name string) error {
	for i, col := range c.Columns {
		if col.ID == id {
			c.Columns[i].Name = name
			return nil
		}
	}
	return fmt.Errorf("colonne %q introuvable", id)
}

// DeleteColumn supprime une colonne par son ID.
func (c *Config) DeleteColumn(id string) error {
	for i, col := range c.Columns {
		if col.ID == id {
			c.Columns = append(c.Columns[:i], c.Columns[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("colonne %q introuvable", id)
}

// MoveColumnLeft déplace une colonne d'une position vers la gauche.
func (c *Config) MoveColumnLeft(id string) error {
	for i, col := range c.Columns {
		if col.ID == id {
			if i == 0 {
				return fmt.Errorf("colonne déjà en première position")
			}
			c.Columns[i], c.Columns[i-1] = c.Columns[i-1], c.Columns[i]
			return nil
		}
	}
	return fmt.Errorf("colonne %q introuvable", id)
}

// MoveColumnRight déplace une colonne d'une position vers la droite.
func (c *Config) MoveColumnRight(id string) error {
	for i, col := range c.Columns {
		if col.ID == id {
			if i == len(c.Columns)-1 {
				return fmt.Errorf("colonne déjà en dernière position")
			}
			c.Columns[i], c.Columns[i+1] = c.Columns[i+1], c.Columns[i]
			return nil
		}
	}
	return fmt.Errorf("colonne %q introuvable", id)
}
