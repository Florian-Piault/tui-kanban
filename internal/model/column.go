package model

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/piflorian/tui-kanban/internal/storage"
	"github.com/piflorian/tui-kanban/internal/styles"
)

type ColumnModel struct {
	ID       string
	Name     string
	Tasks    []storage.Task
	Cursor   int
	scroll   int // index de la première tâche visible
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
		c.scroll = 0
	}
	c.ensureScrollVisible()
}

func (c *ColumnModel) MoveUp() {
	if c.Cursor > 0 {
		c.Cursor--
		c.ensureScrollVisible()
	}
}

func (c *ColumnModel) MoveDown() {
	if c.Cursor < len(c.Tasks)-1 {
		c.Cursor++
		c.ensureScrollVisible()
	}
}

func (c *ColumnModel) SelectedTask() (storage.Task, bool) {
	if len(c.Tasks) == 0 || c.Cursor >= len(c.Tasks) {
		return storage.Task{}, false
	}
	return c.Tasks[c.Cursor], true
}

// cardHeight estime la hauteur en lignes d'une carte (bordures + contenu + marge).
func (c *ColumnModel) cardHeight(task storage.Task) int {
	cardWidth := c.Width - 6 // innerWidth(c.Width-4) - padding(2)
	if cardWidth < 1 {
		cardWidth = 1
	}
	wrapped := styles.WrapText(task.Title, cardWidth)
	titleLines := strings.Count(wrapped, "\n") + 1

	h := 4 + titleLines // border-top(1) + ID(1) + title(n) + border-bottom(1) + margin(1)
	if task.Description != "" {
		h++
	}
	if task.Due != "" {
		h++
	}
	return h
}

// availableH retourne la hauteur disponible pour les cartes dans la colonne.
func (c *ColumnModel) availableH() int {
	h := c.Height - 4 // bordures(2) + titre(1) + séparateur(1)
	if h < 1 {
		h = 1
	}
	return h
}

// ensureScrollVisible ajuste scroll pour que le curseur soit toujours visible.
func (c *ColumnModel) ensureScrollVisible() {
	if len(c.Tasks) == 0 {
		c.scroll = 0
		return
	}
	if c.Cursor < c.scroll {
		c.scroll = c.Cursor
		return
	}
	avail := c.availableH()
	for {
		h := 0
		visible := false
		for i := c.scroll; i < len(c.Tasks); i++ {
			ch := c.cardHeight(c.Tasks[i])
			if h+ch > avail {
				break
			}
			h += ch
			if i == c.Cursor {
				visible = true
				break
			}
		}
		if visible || c.scroll >= c.Cursor {
			break
		}
		c.scroll++
	}
}

func (c ColumnModel) View() string {
	innerWidth := c.Width - 4 // bordure(2) + padding(2)
	if innerWidth < 4 {
		innerWidth = 4
	}

	titleStyle := styles.ColumnTitleStyle
	colStyle := styles.ColumnStyle
	if c.IsActive {
		titleStyle = styles.ColumnTitleActiveStyle
		colStyle = styles.ColumnActiveStyle
	}

	count := fmt.Sprintf(" (%d)", len(c.Tasks))
	colTitle := titleStyle.Render(styles.TruncateTitle(c.Name, innerWidth-utf8.RuneCountInString(count)) + count)

	var lines []string
	lines = append(lines, colTitle)
	lines = append(lines, strings.Repeat("─", innerWidth))

	avail := c.availableH()
	cardWidth := innerWidth - 2

	// Indicateur de défilement supérieur
	if c.scroll > 0 {
		lines = append(lines, styles.HelpStyle.Render(fmt.Sprintf("  ▲ %d au-dessus", c.scroll)))
		avail--
	}

	lastVisible := c.scroll - 1
	usedH := 0

	for i := c.scroll; i < len(c.Tasks); i++ {
		task := c.Tasks[i]
		ch := c.cardHeight(task)

		// Réserver 1 ligne pour l'indicateur inférieur si ce n'est pas la dernière tâche
		reserved := 0
		if i < len(c.Tasks)-1 && usedH+ch >= avail {
			reserved = 1
		}
		if usedH+ch+reserved > avail {
			break
		}

		selected := c.IsActive && i == c.Cursor

		idLine := styles.TaskIDStyle.Render(task.ID)
		titleLine := styles.WrapText(task.Title, cardWidth)
		var descLine string
		if task.Description != "" {
			descLine = lipgloss.NewStyle().
				Foreground(styles.ColorTextDim).
				Italic(true).
				Render(styles.TruncateTitle(task.Description, cardWidth))
		}
		var dueLine string
		if task.Due != "" {
			dueLine = styles.DueStyle.Render("⏰ " + task.Due)
		}

		var cardLines []string
		cardLines = append(cardLines, idLine, titleLine)
		if descLine != "" {
			cardLines = append(cardLines, descLine)
		}
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
		usedH += ch
		lastVisible = i
	}

	// Indicateur de défilement inférieur
	remaining := len(c.Tasks) - lastVisible - 1
	if remaining > 0 {
		lines = append(lines, styles.HelpStyle.Render(fmt.Sprintf("  ▼ %d en-dessous", remaining)))
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
