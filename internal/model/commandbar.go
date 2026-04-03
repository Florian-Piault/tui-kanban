package model

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/piflorian/tui-kanban/internal/command"
	"github.com/piflorian/tui-kanban/internal/styles"
)

type CommandBarModel struct {
	input              textinput.Model
	suggestions        []command.Suggestion
	selectedSuggestion int
	ctx                command.CompletionContext
	Width              int
	err                string
}

func NewCommandBar() CommandBarModel {
	ti := textinput.New()
	ti.Placeholder = "/commande…"
	ti.Focus()
	ti.CharLimit = 256
	ti.SetValue("/")
	ti.CursorEnd()
	return CommandBarModel{input: ti}
}

func (m *CommandBarModel) SetContext(ctx command.CompletionContext) {
	m.ctx = ctx
}

func (m *CommandBarModel) SetWidth(w int) {
	m.Width = w
	m.input.Width = w - 4
}

func (m CommandBarModel) Update(msg tea.Msg) (CommandBarModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if len(m.suggestions) > 0 {
				m.applySelected()
				return m, nil
			}
		case "shift+tab":
			m.navigate(-1)
			return m, nil
		case "up":
			m.navigate(-1)
			return m, nil
		case "down":
			m.navigate(+1)
			return m, nil
		case "esc":
			return m, func() tea.Msg { return CloseCommandBarMsg{} }
		case "enter":
			if len(m.suggestions) > 0 {
				m.applySelected()
				return m, nil
			}
			return m, m.submit()
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			m.refreshSuggestions()
			m.selectedSuggestion = 0
			m.err = ""
			return m, cmd
		}
	}
	return m, nil
}

func (m *CommandBarModel) refreshSuggestions() {
	m.suggestions = command.Complete(m.input.Value(), m.ctx)
}

func (m *CommandBarModel) navigate(delta int) {
	if len(m.suggestions) == 0 {
		return
	}
	m.selectedSuggestion = (m.selectedSuggestion + delta + len(m.suggestions)) % len(m.suggestions)
}

func (m *CommandBarModel) applySelected() {
	if len(m.suggestions) == 0 {
		return
	}
	val := m.suggestions[m.selectedSuggestion].Value
	tokens := command.Tokenize(m.input.Value())
	trailingSpace := len(m.input.Value()) > 1 && m.input.Value()[len(m.input.Value())-1] == ' '

	var newVal string
	if len(tokens) <= 1 && !trailingSpace {
		newVal = "/" + val + " "
	} else {
		if trailingSpace {
			newVal = "/" + strings.Join(tokens, " ") + " " + val + " "
		} else {
			newVal = "/" + strings.Join(tokens[:len(tokens)-1], " ")
			if len(tokens) > 1 {
				newVal += " "
			}
			newVal += val + " "
		}
	}

	m.input.SetValue(newVal)
	m.input.CursorEnd()
	m.refreshSuggestions()
	m.selectedSuggestion = 0
}

func (m CommandBarModel) submit() tea.Cmd {
	val := strings.TrimSpace(m.input.Value())
	if val == "" || val == "/" {
		return func() tea.Msg { return CloseCommandBarMsg{} }
	}
	return func() tea.Msg {
		parsed, err := command.Parse(val)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return CommandParsedMsg{Parsed: parsed}
	}
}

// highlightLabel rend le label avec les caractères matchés en gras/colorés.
func highlightLabel(label string, matchedIndexes []int, selected bool) string {
	const width = 36
	if len(matchedIndexes) == 0 {
		runes := []rune(label)
		pad := width - len(runes)
		if pad < 0 {
			pad = 0
		}
		return label + strings.Repeat(" ", pad)
	}

	idxSet := make(map[int]bool, len(matchedIndexes))
	for _, idx := range matchedIndexes {
		idxSet[idx] = true
	}

	matchFg := styles.ColorPrimary
	if selected {
		matchFg = styles.ColorText
	}
	matchStyle := lipgloss.NewStyle().Bold(true).Foreground(matchFg)

	var sb strings.Builder
	runes := []rune(label)
	for i, r := range runes {
		ch := string(r)
		if idxSet[i] {
			sb.WriteString(matchStyle.Render(ch))
		} else {
			sb.WriteString(ch)
		}
	}
	pad := width - len(runes)
	if pad > 0 {
		sb.WriteString(strings.Repeat(" ", pad))
	}
	return sb.String()
}

func (m CommandBarModel) View() string {
	var sb strings.Builder

	// Dropdown de suggestions (affiché au-dessus)
	if len(m.suggestions) > 0 {
		var suggLines []string
		maxShow := 5
		if len(m.suggestions) < maxShow {
			maxShow = len(m.suggestions)
		}
		start := m.selectedSuggestion - maxShow/2
		if start < 0 {
			start = 0
		}
		if start+maxShow > len(m.suggestions) {
			start = len(m.suggestions) - maxShow
		}
		for i := start; i < start+maxShow; i++ {
			s := m.suggestions[i]
			rendered := highlightLabel(s.Label, s.MatchedIndexes, i == m.selectedSuggestion)
			if i == m.selectedSuggestion {
				suggLines = append(suggLines, styles.SuggestionSelectedStyle.Render(" "+rendered))
			} else {
				suggLines = append(suggLines, styles.SuggestionStyle.Render(" "+rendered))
			}
		}
		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorBorderAct).
			Render(strings.Join(suggLines, "\n"))
		sb.WriteString(box + "\n")
	}

	// Erreur éventuelle
	if m.err != "" {
		sb.WriteString(styles.ErrorStyle.Render("✗ "+m.err) + "\n")
	}

	// Barre de saisie
	bar := styles.CommandBarStyle.Width(m.Width - 2).Render(m.input.View())
	sb.WriteString(bar)

	return sb.String()
}

