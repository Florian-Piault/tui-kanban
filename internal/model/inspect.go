package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/piflorian/tui-kanban/internal/storage"
	"github.com/piflorian/tui-kanban/internal/styles"
)

type inspectSubState int

const (
	inspectBrowsing inspectSubState = iota
	inspectAdding
)

type InspectModel struct {
	task     storage.Task
	cursor   int
	subState inspectSubState
	addInput textinput.Model
	Width    int
}

func NewInspectModel() InspectModel {
	ti := textinput.New()
	ti.Placeholder = "Texte de la sous-tâche…"
	ti.CharLimit = 200
	return InspectModel{addInput: ti}
}

func (m *InspectModel) Open(task storage.Task) {
	m.task = task
	m.cursor = 0
	m.subState = inspectBrowsing
	m.addInput.Reset()
	m.addInput.Blur()
}

func (m InspectModel) Update(msg tea.Msg) (InspectModel, tea.Cmd) {
	if m.subState == inspectAdding {
		return m.updateAdding(msg)
	}
	return m.updateBrowsing(msg)
}

func (m InspectModel) updateBrowsing(msg tea.Msg) (InspectModel, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch keyMsg.String() {
	case "esc":
		return m, func() tea.Msg { return CloseInspectMsg{} }

	case "e":
		task := m.task
		return m, func() tea.Msg {
			return OpenModalMsg{Task: task, IsNew: false, ColID: task.Status}
		}

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.task.Checklist)-1 {
			m.cursor++
		}

	case " ":
		if len(m.task.Checklist) > 0 && m.cursor < len(m.task.Checklist) {
			m.task.Checklist[m.cursor].Done = !m.task.Checklist[m.cursor].Done
			task := m.task
			return m, func() tea.Msg { return checklistUpdatedMsg{task: task} }
		}

	case "a":
		m.subState = inspectAdding
		m.addInput.Reset()
		m.addInput.Focus()
		return m, textinput.Blink

	case "d":
		if len(m.task.Checklist) > 0 && m.cursor < len(m.task.Checklist) {
			m.task.Checklist = append(m.task.Checklist[:m.cursor], m.task.Checklist[m.cursor+1:]...)
			if m.cursor >= len(m.task.Checklist) && m.cursor > 0 {
				m.cursor--
			}
			task := m.task
			return m, func() tea.Msg { return checklistUpdatedMsg{task: task} }
		}
	}
	return m, nil
}

func (m InspectModel) updateAdding(msg tea.Msg) (InspectModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			m.subState = inspectBrowsing
			m.addInput.Blur()
			return m, nil

		case "enter":
			text := strings.TrimSpace(m.addInput.Value())
			m.subState = inspectBrowsing
			m.addInput.Blur()
			m.addInput.Reset()
			if text != "" {
				m.task.Checklist = append(m.task.Checklist, storage.ChecklistItem{Text: text})
				m.cursor = len(m.task.Checklist) - 1
				task := m.task
				return m, func() tea.Msg { return checklistUpdatedMsg{task: task} }
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.addInput, cmd = m.addInput.Update(msg)
	return m, cmd
}

func (m InspectModel) View() string {
	if m.task.ID == "" {
		return ""
	}

	innerWidth := m.Width - 8
	if innerWidth < 20 {
		innerWidth = 20
	}

	var lines []string

	// En-tête : ID + titre
	idStr := lipgloss.NewStyle().Foreground(styles.TypeColor(m.task.Type)).Bold(true).Render(m.task.ID)
	titleStr := lipgloss.NewStyle().Foreground(styles.ColorText).Bold(true).
		Render(styles.TruncateTitle(m.task.Title, innerWidth-len(m.task.ID)-3))
	lines = append(lines, idStr+"  "+titleStr)

	if m.task.Description != "" {
		lines = append(lines, lipgloss.NewStyle().
			Foreground(styles.ColorTextDim).Italic(true).
			Render(styles.TruncateTitle(m.task.Description, innerWidth)))
	}
	if m.task.Due != "" {
		lines = append(lines, styles.DueStyle.Render(m.task.Due))
	}

	lines = append(lines, strings.Repeat("─", innerWidth))

	// Section checklist
	done, total := m.task.ChecklistProgress()
	if total == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(styles.ColorMuted).Italic(true).
			Render("Aucune sous-tâche"))
	} else {
		header := fmt.Sprintf("Checklist  %d/%d", done, total)
		lines = append(lines, lipgloss.NewStyle().Foreground(styles.ColorSecondary).Bold(true).Render(header))

		for i, item := range m.task.Checklist {
			var mark, text string
			if item.Done {
				mark = lipgloss.NewStyle().Foreground(styles.ColorSuccess).Render("[x]")
				text = lipgloss.NewStyle().Foreground(styles.ColorMuted).Strikethrough(true).
					Render(styles.TruncateTitle(item.Text, innerWidth-6))
			} else {
				mark = lipgloss.NewStyle().Foreground(styles.ColorTextDim).Render("[ ]")
				text = lipgloss.NewStyle().Foreground(styles.ColorText).
					Render(styles.TruncateTitle(item.Text, innerWidth-6))
			}

			prefix := "  "
			if m.subState == inspectBrowsing && i == m.cursor {
				prefix = lipgloss.NewStyle().Foreground(styles.ColorPrimary).Bold(true).Render("▶ ")
				text = lipgloss.NewStyle().Foreground(styles.ColorText).Bold(true).
					Render(styles.TruncateTitle(item.Text, innerWidth-6))
				if item.Done {
					text = lipgloss.NewStyle().Foreground(styles.ColorMuted).Strikethrough(true).
						Render(styles.TruncateTitle(item.Text, innerWidth-6))
				}
			}
			lines = append(lines, prefix+mark+" "+text)
		}
	}

	// Zone de saisie d'une nouvelle sous-tâche
	if m.subState == inspectAdding {
		lines = append(lines, strings.Repeat("─", innerWidth))
		prompt := lipgloss.NewStyle().Foreground(styles.ColorSecondary).Render("+ ")
		lines = append(lines, prompt+m.addInput.View())
	}

	lines = append(lines, strings.Repeat("─", innerWidth))

	// Raccourcis
	var hints string
	if m.subState == inspectAdding {
		hints = "entrée: confirmer  •  esc: annuler"
	} else {
		parts := []string{"espace: toggle", "a: ajouter", "e: éditer", "esc: fermer"}
		if total > 0 {
			parts = append(parts[:2], append([]string{"d: supprimer"}, parts[2:]...)...)
		}
		hints = strings.Join(parts, "  •  ")
	}
	lines = append(lines, styles.HelpStyle.Render(hints))

	content := strings.Join(lines, "\n")
	width := m.Width - 4
	if width < 40 {
		width = 40
	}
	return styles.ModalStyle.Width(width).Render(content)
}
