package model

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/piflorian/tui-kanban/internal/config"
	"github.com/piflorian/tui-kanban/internal/storage"
	"github.com/piflorian/tui-kanban/internal/styles"
)

type BoardModel struct {
	Columns   []ColumnModel
	ActiveCol int
	Width     int
	Height    int
}

func NewBoard(cfg *config.Config) BoardModel {
	cols := make([]ColumnModel, len(cfg.Columns))
	for i, c := range cfg.Columns {
		cols[i] = NewColumn(c.ID, c.Name)
	}
	if len(cols) > 0 {
		cols[0].IsActive = true
	}
	return BoardModel{Columns: cols}
}

func (b *BoardModel) SetSize(width, height int) {
	b.Width = width
	b.Height = height

	widths := styles.ComputeColumnWidths(width, len(b.Columns))
	for i := range b.Columns {
		b.Columns[i].Width = widths[i]
		b.Columns[i].Height = height
	}
}

func (b *BoardModel) SetTasks(colID string, tasks []storage.Task) {
	for i := range b.Columns {
		if b.Columns[i].ID == colID {
			b.Columns[i].SetTasks(tasks)
			return
		}
	}
}

// SetColumnTotal met à jour le total non filtré d'une colonne (pour afficher X/Y).
func (b *BoardModel) SetColumnTotal(colID string, total int) {
	for i := range b.Columns {
		if b.Columns[i].ID == colID {
			b.Columns[i].AllTasksCount = total
			return
		}
	}
}

// SetFilter propage le filtre actif à toutes les colonnes (pour le highlighting).
func (b *BoardModel) SetFilter(q FilterQuery) {
	for i := range b.Columns {
		b.Columns[i].ActiveFilter = q
	}
}

func (b *BoardModel) ActiveColumn() *ColumnModel {
	if b.ActiveCol >= 0 && b.ActiveCol < len(b.Columns) {
		return &b.Columns[b.ActiveCol]
	}
	return nil
}

func (b *BoardModel) ActiveColumnID() string {
	col := b.ActiveColumn()
	if col == nil {
		return ""
	}
	return col.ID
}

func (b *BoardModel) SelectedTask() (storage.Task, bool) {
	col := b.ActiveColumn()
	if col == nil {
		return storage.Task{}, false
	}
	return col.SelectedTask()
}

func (b *BoardModel) MoveLeft() {
	if b.ActiveCol > 0 {
		b.Columns[b.ActiveCol].IsActive = false
		b.ActiveCol--
		b.Columns[b.ActiveCol].IsActive = true
	}
}

func (b *BoardModel) MoveRight() {
	if b.ActiveCol < len(b.Columns)-1 {
		b.Columns[b.ActiveCol].IsActive = false
		b.ActiveCol++
		b.Columns[b.ActiveCol].IsActive = true
	}
}

func (b *BoardModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "h", "left":
			b.MoveLeft()
		case "l", "right":
			b.MoveRight()
		case "j", "down":
			col := b.ActiveColumn()
			if col != nil {
				col.MoveDown()
			}
		case "k", "up":
			col := b.ActiveColumn()
			if col != nil {
				col.MoveUp()
			}
		}
	case TasksLoadedMsg:
		b.SetTasks(msg.ColID, msg.Tasks)
	}
	return nil
}

func (b BoardModel) View() string {
	return b.ViewAtHeight(b.Height)
}

// ViewAtHeight rend le board avec une hauteur spécifique sans muter l'état.
// Les colonnes étant des valeurs (pas des pointeurs), col est une copie locale.
func (b BoardModel) ViewAtHeight(h int) string {
	if len(b.Columns) == 0 {
		return lipgloss.NewStyle().
			Foreground(styles.ColorMuted).
			Render("Aucune colonne configurée.")
	}

	widths := styles.ComputeColumnWidths(b.Width, len(b.Columns))
	cols := make([]string, len(b.Columns))
	for i, col := range b.Columns {
		col.Height = h
		col.Width = widths[i]
		cols[i] = col.View()
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}
