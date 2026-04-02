package model

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/piflorian/tui-kanban/internal/storage"
	"github.com/piflorian/tui-kanban/internal/styles"
)

type modalField int

const (
	fieldTitle modalField = iota
	fieldDescription
	fieldDue
	fieldCount
)

type ModalModel struct {
	fields  [fieldCount]textinput.Model
	focused modalField
	task    storage.Task
	isNew   bool
	colID   string
	Width   int
}

func NewModal() ModalModel {
	labels := []string{"Titre", "Description", "Échéance"}
	placeholders := []string{"Titre de la tâche", "Description (optionnelle)", "YYYY-MM-DD (optionnel)"}

	var fields [fieldCount]textinput.Model
	for i := range fields {
		ti := textinput.New()
		ti.Placeholder = placeholders[i]
		ti.CharLimit = 256
		_ = labels[i]
		fields[i] = ti
	}
	fields[fieldTitle].Focus()

	return ModalModel{fields: fields, isNew: true}
}

func (m *ModalModel) Open(task storage.Task, isNew bool, colID string) {
	m.task = task
	m.isNew = isNew
	m.colID = colID
	m.focused = fieldTitle

	m.fields[fieldTitle].SetValue(task.Title)
	m.fields[fieldDescription].SetValue(task.Description)
	m.fields[fieldDue].SetValue(task.Due)

	for i := range m.fields {
		m.fields[i].Blur()
	}
	m.fields[fieldTitle].Focus()
}

func (m ModalModel) Update(msg tea.Msg) (ModalModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return CloseModalMsg{} }
		case "tab", "down":
			m.fields[m.focused].Blur()
			m.focused = (m.focused + 1) % fieldCount
			m.fields[m.focused].Focus()
			return m, nil
		case "shift+tab", "up":
			m.fields[m.focused].Blur()
			m.focused = (m.focused - 1 + fieldCount) % fieldCount
			m.fields[m.focused].Focus()
			return m, nil
		case "ctrl+s", "enter":
			if m.focused == fieldDue || msg.String() == "ctrl+s" {
				return m, m.submit()
			}
			// enter dans un champ intermédiaire → champ suivant
			m.fields[m.focused].Blur()
			m.focused = (m.focused + 1) % fieldCount
			m.fields[m.focused].Focus()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.fields[m.focused], cmd = m.fields[m.focused].Update(msg)
	return m, cmd
}

func (m ModalModel) submit() tea.Cmd {
	task := m.task
	task.Title = strings.TrimSpace(m.fields[fieldTitle].Value())
	task.Description = strings.TrimSpace(m.fields[fieldDescription].Value())
	task.Due = strings.TrimSpace(m.fields[fieldDue].Value())

	if m.isNew {
		task.Status = m.colID
	}

	isNew := m.isNew
	return func() tea.Msg {
		if isNew {
			return TaskCreatedMsg{Task: task}
		}
		return TaskUpdatedMsg{Task: task}
	}
}

func (m ModalModel) View() string {
	title := "Nouvelle tâche"
	if !m.isNew {
		title = "Modifier " + m.task.ID
	}

	labels := []string{"Titre       ", "Description ", "Échéance    "}
	hints := []string{"", "", "format : YYYY-MM-DD"}

	var lines []string
	lines = append(lines, styles.ModalTitleStyle.Render(title))
	lines = append(lines, "")

	for i, field := range m.fields {
		label := styles.LabelStyle.Render(labels[i]+":")
		input := field.View()
		line := lipgloss.JoinHorizontal(lipgloss.Top, label, " ", input)
		lines = append(lines, line)
		if hints[i] != "" {
			lines = append(lines, lipgloss.NewStyle().
				Foreground(styles.ColorMuted).
				MarginLeft(14).
				Render(hints[i]))
		}
		lines = append(lines, "")
	}

	lines = append(lines, styles.HelpStyle.Render("Tab/Shift+Tab : naviguer  •  Ctrl+S : valider  •  Esc : annuler"))

	content := strings.Join(lines, "\n")
	width := m.Width - 8
	if width < 40 {
		width = 40
	}

	return styles.ModalStyle.Width(width).Render(content)
}
