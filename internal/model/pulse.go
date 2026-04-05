package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/piflorian/tui-kanban/internal/storage"
	"github.com/piflorian/tui-kanban/internal/styles"
)

// PulseDay représente l'activité d'une journée.
type PulseDay struct {
	Date  time.Time
	Count int
}

// PulseModel gère la vue graphe de vélocité et le mode Zen.
type PulseModel struct {
	days       [7]PulseDay
	maxCount   int
	totalCount int
	width      int
	height     int
	zenTask    storage.Task
}

func NewPulseModel() PulseModel {
	return PulseModel{}
}

func (p *PulseModel) SetSize(w, h int) {
	p.width = w
	p.height = h
}

// Compute calcule les counts sur les 7 derniers jours depuis allTasks.
// Si Updated est zéro sur une tâche, on utilise Created en fallback.
func (p *PulseModel) Compute(allTasks map[string][]storage.Task) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Initialiser les 7 jours (du plus ancien au plus récent)
	for i := 0; i < 7; i++ {
		p.days[i] = PulseDay{Date: today.AddDate(0, 0, i-6), Count: 0}
	}

	// Compter les tâches par jour
	for _, tasks := range allTasks {
		for _, t := range tasks {
			ref := t.Updated
			if ref.IsZero() {
				ref = t.Created
			}
			if ref.IsZero() {
				continue
			}
			day := time.Date(ref.Year(), ref.Month(), ref.Day(), 0, 0, 0, 0, ref.Location())
			for i := range p.days {
				if p.days[i].Date.Equal(day) {
					p.days[i].Count++
					break
				}
			}
		}
	}

	// Calculer max et total
	p.maxCount = 0
	p.totalCount = 0
	for _, d := range p.days {
		p.totalCount += d.Count
		if d.Count > p.maxCount {
			p.maxCount = d.Count
		}
	}
}

// barChar retourne le caractère Unicode correspondant à l'intensité.
func barChar(count, maxCount int) rune {
	blocks := []rune{'░', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	if count == 0 {
		return blocks[0]
	}
	if maxCount == 0 {
		return blocks[1]
	}
	idx := (count * 8) / (maxCount + 1)
	if idx > 8 {
		idx = 8
	}
	if idx < 1 {
		idx = 1
	}
	return blocks[idx]
}

// dayLabel retourne le label court d'une date en français.
func dayLabel(t time.Time) string {
	days := []string{"dim", "lun", "mar", "mer", "jeu", "ven", "sam"}
	return fmt.Sprintf("%s %02d", days[t.Weekday()], t.Day())
}

// View affiche le graphe complet des 7 derniers jours.
func (p PulseModel) View() string {
	titleStyle := lipgloss.NewStyle().Foreground(styles.ColorSecondary).Bold(true).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)
	hintStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted).Padding(0, 1)

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Project Pulse — 7 derniers jours"))
	sb.WriteString("\n\n")

	for _, d := range p.days {
		bar := barChar(d.Count, p.maxCount)
		color := styles.PulseBarColor(d.Count)
		barStyled := lipgloss.NewStyle().Foreground(color).Render(string(bar))
		label := dimStyle.Render("  " + dayLabel(d.Date))
		count := dimStyle.Render(fmt.Sprintf("  %2d", d.Count))
		sb.WriteString(label + "  " + barStyled + count + "\n")
	}

	sb.WriteString("\n")

	var avg float64
	if p.totalCount > 0 {
		avg = float64(p.totalCount) / 7.0
	}
	summary := fmt.Sprintf("  Total : %d tâches  •  Moy : %.1f/jour", p.totalCount, avg)
	sb.WriteString(dimStyle.Render(summary))
	sb.WriteString("\n")
	sb.WriteString(hintStyle.Render("tab / esc : retour au board"))

	return sb.String()
}

// ViewZen affiche la tâche sélectionnée en plein écran avec un mini-graphe.
func (p PulseModel) ViewZen() string {
	task := p.zenTask
	maxWidth := 60
	if p.width > 0 && p.width < maxWidth {
		maxWidth = p.width - 4
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(1, 2).
		Width(maxWidth)

	titleStyle := lipgloss.NewStyle().Foreground(styles.ColorText).Bold(true)
	idStyle := lipgloss.NewStyle().Foreground(styles.TypeColor(task.Type)).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(styles.ColorTextDim).Italic(true)
	hintStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)

	var cardContent strings.Builder
	cardContent.WriteString(idStyle.Render(task.ID) + "  " + titleStyle.Render(task.Title))

	if task.Description != "" {
		cardContent.WriteString("\n\n")
		cardContent.WriteString(dimStyle.Render(task.Description))
	}

	done, total := task.ChecklistProgress()
	if total > 0 {
		cardContent.WriteString("\n\n")
		bar := buildProgressBar(done, total, 20)
		cardContent.WriteString(fmt.Sprintf("%s  %d/%d items", bar, done, total))
	}

	card := cardStyle.Render(cardContent.String())

	// Mini graphe sur une ligne
	var miniBar strings.Builder
	miniBar.WriteString("  ")
	for _, d := range p.days {
		bar := barChar(d.Count, p.maxCount)
		color := styles.PulseBarColor(d.Count)
		miniBar.WriteString(lipgloss.NewStyle().Foreground(color).Render(string(bar)))
	}

	hint := hintStyle.Render("  z / esc : quitter zen  •  e : éditer")

	// Centrer la carte
	var sb strings.Builder
	leftPad := 0
	if p.width > maxWidth+4 {
		leftPad = (p.width - maxWidth - 4) / 2
	}
	pad := strings.Repeat(" ", leftPad)

	lines := strings.Split(card, "\n")
	for _, l := range lines {
		sb.WriteString(pad + l + "\n")
	}
	sb.WriteString("\n")
	sb.WriteString(miniBar.String() + "\n")
	sb.WriteString(hint)

	return sb.String()
}

// buildProgressBar construit une barre de progression textuelle.
func buildProgressBar(done, total, width int) string {
	if total == 0 {
		return ""
	}
	filled := (done * width) / total
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	color := styles.PulseBarColor(done * 7 / total)
	return lipgloss.NewStyle().Foreground(color).Render(bar)
}
