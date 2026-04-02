package model

import (
	"strconv"
	"strings"
	"time"

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
	placeholders := []string{"Titre de la tâche", "Description (optionnelle)", "YYYY-MM-DD (optionnel)"}

	var fields [fieldCount]textinput.Model
	for i := range fields {
		ti := textinput.New()
		ti.Placeholder = placeholders[i]
		ti.CharLimit = 256
		fields[i] = ti
	}
	fields[fieldDescription].CharLimit = 0
	fields[fieldDue].CharLimit = 10
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
	m.applyInputWidths()
}

// labelWidth = len("Description :") + 1 espace séparateur = 14
const modalLabelWidth = 14

func (m *ModalModel) applyInputWidths() {
	inputWidth := m.Width - 8 - modalLabelWidth
	if inputWidth < 20 {
		inputWidth = 20
	}
	for i := range m.fields {
		m.fields[i].Width = inputWidth
	}
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
			m.fields[m.focused].Blur()
			m.focused = (m.focused + 1) % fieldCount
			m.fields[m.focused].Focus()
			return m, nil
		case "backspace", "ctrl+h":
			if m.focused == fieldDue {
				m.fields[fieldDue].SetValue(dueBackspace(m.fields[fieldDue].Value()))
				m.fields[fieldDue].CursorEnd()
				return m, nil
			}
		default:
			if m.focused == fieldDue {
				s := msg.String()
				if len(s) == 1 && s[0] >= '0' && s[0] <= '9' {
					m.fields[fieldDue].SetValue(dueInsertDigit(m.fields[fieldDue].Value(), s))
					m.fields[fieldDue].CursorEnd()
				}
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.fields[m.focused], cmd = m.fields[m.focused].Update(msg)
	return m, cmd
}

// dueInsertDigit ajoute un chiffre validé à la valeur YYYY-MM-DD en cours de saisie.
func dueInsertDigit(val, digit string) string {
	digits := strings.ReplaceAll(val, "-", "")
	if len(digits) >= 8 {
		return val
	}
	d := digit[0]
	switch len(digits) {
	case 4: // 1er chiffre du mois : 0 ou 1
		if d != '0' && d != '1' {
			return val
		}
	case 5: // 2e chiffre du mois : mois 01-12
		m1 := digits[4]
		if m1 == '0' && d == '0' { // 00 invalide
			return val
		}
		if m1 == '1' && d > '2' { // 13-19 invalide
			return val
		}
	case 6: // 1er chiffre du jour : 0-3, mais 3 interdit si le mois a < 30 jours (ex: février)
		if d < '0' || d > '3' {
			return val
		}
		if d == '3' {
			year, _ := strconv.Atoi(digits[0:4])
			month, _ := strconv.Atoi(digits[4:6])
			if daysInMonth(year, month) < 30 {
				return val
			}
		}
	case 7: // 2e chiffre du jour : valider contre le vrai mois
		d1 := digits[6]
		if d1 == '0' && d == '0' { // 00 invalide
			return val
		}
		year, _ := strconv.Atoi(digits[0:4])
		month, _ := strconv.Atoi(digits[4:6])
		day := int(d1-'0')*10 + int(d-'0')
		if day > daysInMonth(year, month) {
			return val
		}
	}
	return formatDueDigits(digits + digit)
}

// dueBackspace supprime le dernier chiffre saisi.
func dueBackspace(val string) string {
	digits := strings.ReplaceAll(val, "-", "")
	if len(digits) == 0 {
		return ""
	}
	return formatDueDigits(digits[:len(digits)-1])
}

// daysInMonth retourne le nombre de jours dans un mois donné (gère les années bissextiles).
func daysInMonth(year, month int) int {
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()
}

// formatDueDigits reconstruit YYYY-MM-DD et affiche le '-' dès qu'on atteint la position de séparation.
func formatDueDigits(digits string) string {
	var sb strings.Builder
	for i, c := range digits {
		if i == 4 || i == 6 {
			sb.WriteByte('-')
		}
		sb.WriteRune(c)
	}
	// Affiche le tiret de séparation dès qu'on a exactement 4 ou 6 chiffres
	n := len(digits)
	if n == 4 || n == 6 {
		sb.WriteByte('-')
	}
	return sb.String()
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
	hints := []string{"", "", ""}

	var lines []string
	lines = append(lines, styles.ModalTitleStyle.Render(title))
	lines = append(lines, "")

	for i, field := range m.fields {
		label := styles.LabelStyle.Render(labels[i] + ":")
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
