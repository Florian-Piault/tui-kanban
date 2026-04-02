package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/piflorian/tui-kanban/internal/storage"
	"github.com/piflorian/tui-kanban/internal/styles"
)

type ColumnModel struct {
	ID       string
	Name     string
	Tasks    []storage.Task
	Cursor   int
	Width    int
	Height   int
	IsActive bool
}

func NewColumn(id, name string) ColumnModel {
	return ColumnModel{
		ID:   id,
		Name: name,
	}
}

func (c *ColumnModel) SetTasks(tasks []storage.Task) {
	c.Tasks = tasks
	if c.Cursor >= len(tasks) {
		c.Cursor = max(0, len(tasks)-1)
	}
}

func (c *ColumnModel) MoveUp() {
	if c.Cursor > 0 {
		c.Cursor--
	}
}

func (c *ColumnModel) MoveDown() {
	if c.Cursor < len(c.Tasks)-1 {
		c.Cursor++
	}
}

func (c *ColumnModel) SelectedTask() (storage.Task, bool) {
	if len(c.Tasks) == 0 || c.Cursor >= len(c.Tasks) {
		return storage.Task{}, false
	}
	return c.Tasks[c.Cursor], true
}

func (c ColumnModel) View() string {
	innerWidth := c.Width - 4 // bordure (2) + padding (2)
	if innerWidth < 4 {
		innerWidth = 4
	}

	// En-tête
	titleStyle := styles.ColumnTitleStyle
	colStyle := styles.ColumnStyle
	if c.IsActive {
		titleStyle = styles.ColumnTitleActiveStyle
		colStyle = styles.ColumnActiveStyle
	}

	count := fmt.Sprintf(" (%d)", len(c.Tasks))
	title := titleStyle.Render(styles.TruncateTitle(c.Name, innerWidth-len(count)) + count)

	// Corps : liste des cartes
	var lines []string
	lines = append(lines, title)
	lines = append(lines, strings.Repeat("─", innerWidth))

	maxCards := c.Height - 4 // titre + séparateur + bordures
	if maxCards < 1 {
		maxCards = 1
	}

	for i, task := range c.Tasks {
		if i >= maxCards {
			more := styles.HelpStyle.Render(fmt.Sprintf("  +%d autres…", len(c.Tasks)-maxCards))
			lines = append(lines, more)
			break
		}

		cardWidth := innerWidth - 2
		selected := c.IsActive && i == c.Cursor

		titleLine := styles.TruncateTitle(task.Title, cardWidth)
		var dueLine string
		if task.Due != "" {
			dueLine = styles.DueStyle.Render("⏰ " + task.Due)
		}

		var cardLines []string
		cardLines = append(cardLines, titleLine)
		if dueLine != "" {
			cardLines = append(cardLines, dueLine)
		}
		cardContent := strings.Join(cardLines, "\n")

		var card string
		if selected {
			card = styles.CardSelectedStyle.Width(cardWidth).Render(cardContent)
		} else {
			card = styles.CardStyle.Width(cardWidth).Render(cardContent)
		}
		lines = append(lines, card)
	}

	if len(c.Tasks) == 0 {
		empty := lipgloss.NewStyle().
			Foreground(styles.ColorMuted).
			Italic(true).
			Padding(1, 0).
			Render("Vide")
		lines = append(lines, empty)
	}

	body := strings.Join(lines, "\n")
	return colStyle.Width(c.Width - 2).Height(c.Height - 2).Render(body)
}

