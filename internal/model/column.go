package model

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/piflorian/tui-kanban/internal/storage"
	"github.com/piflorian/tui-kanban/internal/styles"
)

// renderChecklistBar retourne une ligne de progression compacte pour les cartes.
func renderChecklistBar(done, total int) string {
	const barWidth = 6
	filled := barWidth * done / total
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	color := styles.ColorPrimary
	if done == total {
		color = styles.ColorSuccess
	}
	barStr := lipgloss.NewStyle().Foreground(color).Render(bar)
	return fmt.Sprintf("%d/%d ", done, total) + barStr
}

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
	wrapped := styles.WrapText(task.Title, cardWidth-2) // lipgloss wraps à width-leftPad-rightPad
	titleLines := strings.Count(wrapped, "\n") + 1

	h := 4 + titleLines // border-top(1) + ID(1) + title(n) + border-bottom(1) + margin(1)
	if task.Description != "" {
		h++
	}
	if task.Due != "" {
		h++
	}
	if _, total := task.ChecklistProgress(); total > 0 {
		h++
	}
	return h
}

// cardAreaHeight retourne la hauteur disponible pour les cartes.
// Overhead fixe : bordures(2) + titre(1) + séparateur(1) + indicateur_haut(1) + indicateur_bas(1) = 6
func (c *ColumnModel) cardAreaHeight() int {
	h := c.Height - 6
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
	avail := c.cardAreaHeight()
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

	// Indicateur supérieur : toujours rendu (ligne vide si pas de scroll).
	if c.scroll > 0 {
		lines = append(lines, styles.HelpStyle.Render(fmt.Sprintf("  ▲ %d au-dessus", c.scroll)))
	} else {
		lines = append(lines, "")
	}

	avail := c.cardAreaHeight()
	cardWidth := innerWidth - 2

	if len(c.Tasks) == 0 {
		empty := lipgloss.NewStyle().
			Foreground(styles.ColorMuted).
			Italic(true).
			Render("Vide")
		lines = append(lines, empty)
		// Indicateur bas : ligne vide pour l'alignement
		lines = append(lines, "")
	} else {
		lastVisible := c.scroll - 1
		usedH := 0

		for i := c.scroll; i < len(c.Tasks); i++ {
			task := c.Tasks[i]
			ch := c.cardHeight(task)

			if usedH+ch > avail {
				break
			}

			selected := c.IsActive && i == c.Cursor

			idLine := lipgloss.NewStyle().Foreground(styles.TypeColor(task.Type)).Bold(true).Render(task.ID)
			titleLine := styles.WrapText(task.Title, cardWidth-2)
			var descLine string
			if task.Description != "" {
				descLine = lipgloss.NewStyle().
					Foreground(styles.ColorTextDim).
					Italic(true).
					Render(styles.TruncateTitle(task.Description, cardWidth))
			}
			var dueLine string
			if task.Due != "" {
				dueLine = styles.DueStyle.Render(task.Due)
			}

			done, total := task.ChecklistProgress()
			var progressLine string
			if total > 0 {
				progressLine = renderChecklistBar(done, total)
			}

			var cardLines []string
			cardLines = append(cardLines, idLine, titleLine)
			if descLine != "" {
				cardLines = append(cardLines, descLine)
			}
			if dueLine != "" {
				cardLines = append(cardLines, dueLine)
			}
			if progressLine != "" {
				cardLines = append(cardLines, progressLine)
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

		// Indicateur inférieur : toujours rendu (ligne vide si tout est visible).
		remaining := len(c.Tasks) - lastVisible - 1
		if remaining > 0 {
			lines = append(lines, styles.HelpStyle.Render(fmt.Sprintf("  ▼ %d en-dessous", remaining)))
		} else {
			lines = append(lines, "")
		}
	}

	body := strings.Join(lines, "\n")
	return colStyle.Width(c.Width - 2).Height(c.Height - 2).Render(body)
}
