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

// focusTarget représente le champ actif dans le modal.
type focusTarget int

const (
	ftTitle       focusTarget = iota // 0 — textinput[0]
	ftType                           // 1 — sélecteur cyclique
	ftDescription                    // 2 — textinput[1]
	ftDue                            // 3 — textinput[2]
	ftCount                          // 4
)

// textInputIdx mappe un focus vers son indice dans inputs[] (-1 si pas un textinput).
func textInputIdx(f focusTarget) int {
	switch f {
	case ftTitle:
		return 0
	case ftDescription:
		return 1
	case ftDue:
		return 2
	}
	return -1
}

type ModalModel struct {
	inputs   [3]textinput.Model // title=0, description=1, due=2
	focused  focusTarget
	taskType string
	task     storage.Task
	isNew    bool
	colID    string
	Width    int
}

func NewModal() ModalModel {
	placeholders := []string{"Titre de la tâche", "Description (optionnelle)", "YYYY-MM-DD (optionnel)"}

	var inputs [3]textinput.Model
	for i := range inputs {
		ti := textinput.New()
		ti.Placeholder = placeholders[i]
		ti.CharLimit = 256
		inputs[i] = ti
	}
	inputs[1].CharLimit = 0 // description illimitée
	inputs[2].CharLimit = 10
	inputs[0].Focus()

	return ModalModel{inputs: inputs, isNew: true, taskType: storage.TypeTask}
}

func (m *ModalModel) Open(task storage.Task, isNew bool, colID string) {
	m.task = task
	m.isNew = isNew
	m.colID = colID
	m.focused = ftTitle

	if task.Type != "" {
		m.taskType = storage.NormalizeType(task.Type)
	} else {
		m.taskType = storage.TypeTask
	}

	m.inputs[0].SetValue(task.Title)
	m.inputs[1].SetValue(task.Description)
	m.inputs[2].SetValue(task.Due)

	for i := range m.inputs {
		m.inputs[i].Blur()
	}
	m.inputs[0].Focus()
	m.applyInputWidths()
}

// labelWidth = len("Description :") + 1 espace séparateur = 14
const modalLabelWidth = 14

func (m *ModalModel) applyInputWidths() {
	inputWidth := m.Width - 8 - modalLabelWidth
	if inputWidth < 20 {
		inputWidth = 20
	}
	for i := range m.inputs {
		m.inputs[i].Width = inputWidth
	}
}

func (m ModalModel) Update(msg tea.Msg) (ModalModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return CloseModalMsg{} }

		case "tab", "down":
			if idx := textInputIdx(m.focused); idx >= 0 {
				m.inputs[idx].Blur()
			}
			m.focused = (m.focused + 1) % ftCount
			if idx := textInputIdx(m.focused); idx >= 0 {
				m.inputs[idx].Focus()
			}
			return m, nil

		case "shift+tab", "up":
			if idx := textInputIdx(m.focused); idx >= 0 {
				m.inputs[idx].Blur()
			}
			m.focused = (m.focused - 1 + ftCount) % ftCount
			if idx := textInputIdx(m.focused); idx >= 0 {
				m.inputs[idx].Focus()
			}
			return m, nil

		case "ctrl+s", "enter":
			if m.focused == ftDue || msg.String() == "ctrl+s" {
				return m, m.submit()
			}
			if m.focused == ftType {
				// Enter sur le sélecteur de type → avancer au champ suivant
				if idx := textInputIdx(m.focused); idx >= 0 {
					m.inputs[idx].Blur()
				}
				m.focused = (m.focused + 1) % ftCount
				if idx := textInputIdx(m.focused); idx >= 0 {
					m.inputs[idx].Focus()
				}
				return m, nil
			}
			if idx := textInputIdx(m.focused); idx >= 0 {
				m.inputs[idx].Blur()
			}
			m.focused = (m.focused + 1) % ftCount
			if idx := textInputIdx(m.focused); idx >= 0 {
				m.inputs[idx].Focus()
			}
			return m, nil

		case " ", "right", "l":
			if m.focused == ftType {
				m.taskType = nextTaskType(m.taskType)
				return m, nil
			}

		case "left", "h":
			if m.focused == ftType {
				m.taskType = prevTaskType(m.taskType)
				return m, nil
			}

		case "backspace", "ctrl+h":
			if m.focused == ftDue {
				m.inputs[2].SetValue(dueBackspace(m.inputs[2].Value()))
				m.inputs[2].CursorEnd()
				return m, nil
			}

		default:
			if m.focused == ftDue {
				s := msg.String()
				if len(s) == 1 && s[0] >= '0' && s[0] <= '9' {
					m.inputs[2].SetValue(dueInsertDigit(m.inputs[2].Value(), s))
					m.inputs[2].CursorEnd()
				}
				return m, nil
			}
		}
	}

	if idx := textInputIdx(m.focused); idx >= 0 {
		var cmd tea.Cmd
		m.inputs[idx], cmd = m.inputs[idx].Update(msg)
		return m, cmd
	}
	return m, nil
}

func nextTaskType(current string) string {
	for i, t := range storage.AllTypes {
		if t == current {
			return storage.AllTypes[(i+1)%len(storage.AllTypes)]
		}
	}
	return storage.TypeTask
}

func prevTaskType(current string) string {
	for i, t := range storage.AllTypes {
		if t == current {
			return storage.AllTypes[(i-1+len(storage.AllTypes))%len(storage.AllTypes)]
		}
	}
	return storage.TypeTask
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
	n := len(digits)
	if n == 4 || n == 6 {
		sb.WriteByte('-')
	}
	return sb.String()
}

func (m ModalModel) submit() tea.Cmd {
	task := m.task
	task.Title = strings.TrimSpace(m.inputs[0].Value())
	task.Type = m.taskType
	task.Description = strings.TrimSpace(m.inputs[1].Value())
	task.Due = strings.TrimSpace(m.inputs[2].Value())

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

	labels := []string{"Titre       ", "Type        ", "Description ", "Échéance    "}

	var lines []string
	lines = append(lines, styles.ModalTitleStyle.Render(title))
	lines = append(lines, "")

	for i := focusTarget(0); i < ftCount; i++ {
		label := styles.LabelStyle.Render(labels[i] + ":")

		var content string
		switch i {
		case ftType:
			content = renderTypeSelector(m.taskType, m.focused == ftType)
		default:
			idx := textInputIdx(i)
			content = m.inputs[idx].View()
		}

		line := lipgloss.JoinHorizontal(lipgloss.Top, label, " ", content)
		lines = append(lines, line)
		lines = append(lines, "")
	}

	hint := "Tab/↑↓ : naviguer  •  ◄► : changer le type  •  Ctrl+S : valider  •  Esc : annuler"
	lines = append(lines, styles.HelpStyle.Render(hint))

	content := strings.Join(lines, "\n")
	width := m.Width - 8
	if width < 40 {
		width = 40
	}

	return styles.ModalStyle.Width(width).Render(content)
}

// renderTypeSelector affiche le sélecteur de type avec badge coloré.
func renderTypeSelector(taskType string, focused bool) string {
	color := styles.TypeColor(taskType)
	label := strings.ToUpper(taskType)

	badgeStyle := lipgloss.NewStyle().Foreground(color).Bold(true)

	if focused {
		arrow := lipgloss.NewStyle().Foreground(styles.ColorTextDim).Render
		return arrow("◄ ") + badgeStyle.Render(label) + arrow(" ►")
	}
	return badgeStyle.Render(label)
}
