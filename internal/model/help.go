package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/piflorian/tui-kanban/internal/styles"
)

// CloseHelpMsg est émis quand l'utilisateur ferme la modale d'aide.
type CloseHelpMsg struct{}

// HelpModel est une modale scrollable affichant l'aide complète de l'application.
type HelpModel struct {
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}

func NewHelpModel() HelpModel { return HelpModel{} }

// Open initialise (ou réinitialise) le viewport avec les dimensions courantes.
func (m *HelpModel) Open(width, height int) {
	m.width = width
	m.height = height

	modalW := width - 8
	if modalW < 40 {
		modalW = 40
	}
	modalH := height - 4
	if modalH < 10 {
		modalH = 10
	}
	// overhead ModalStyle : border top/bot(2) + padding top/bot(2) = 4 lignes
	// lignes fixes internes : titre(1) + sep haut(1) + sep bas(1) + scrollLine(1) = 4
	vpH := modalH - 8
	if vpH < 5 {
		vpH = 5
	}
	// overhead horizontal ModalStyle : border(2) + padding(2*2) = 6 → vpW = modalW - 4 (padding seul)
	vpW := modalW - 4
	if vpW < 20 {
		vpW = 20
	}

	vp := viewport.New(vpW, vpH)
	vp.SetContent(helpContent(vpW))
	vp.GotoTop()
	m.viewport = vp
	m.ready = true
}

func (m HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd) {
	if !m.ready {
		return m, nil
	}

	key, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	switch key.String() {
	case "q", "esc":
		return m, func() tea.Msg { return CloseHelpMsg{} }
	case "g":
		m.viewport.GotoTop()
		return m, nil
	case "G":
		m.viewport.GotoBottom()
		return m, nil
	default:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
}

func (m HelpModel) View() string {
	if !m.ready {
		return ""
	}

	modalW := m.width - 8
	if modalW < 40 {
		modalW = 40
	}
	vpW := modalW - 4

	title := styles.ModalTitleStyle.Render("  Aide — tui-kanban")
	sep := lipgloss.NewStyle().Foreground(styles.ColorBorder).Render(strings.Repeat("─", vpW))
	content := m.viewport.View()

	totalLines := m.viewport.TotalLineCount()
	currentLine := m.viewport.YOffset + m.viewport.Height
	if currentLine > totalLines {
		currentLine = totalLines
	}
	pct := int(m.viewport.ScrollPercent() * 100)

	left := styles.HelpStyle.Render("j/k : défiler  •  g/G : début/fin  •  q/Esc : fermer")
	right := lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
		fmt.Sprintf("%d/%d  %d%%", currentLine, totalLines, pct))
	lw := lipgloss.Width(left)
	rw := lipgloss.Width(right)
	space := vpW - lw - rw
	if space < 1 {
		space = 1
	}
	scrollLine := left + strings.Repeat(" ", space) + right

	inner := strings.Join([]string{title, sep, content, sep, scrollLine}, "\n")
	return styles.ModalStyle.Width(modalW).Render(inner)
}

// helpContent génère le contenu d'aide structuré avec mise en forme lipgloss.
func helpContent(width int) string {
	h := func(title string) string {
		return lipgloss.NewStyle().Foreground(styles.ColorSecondary).Bold(true).Render(title)
	}
	dim := func(s string) string {
		return lipgloss.NewStyle().Foreground(styles.ColorTextDim).Render(s)
	}
	key := func(s string) string {
		return lipgloss.NewStyle().Foreground(styles.ColorPrimary).Bold(true).Render(s)
	}
	sep := lipgloss.NewStyle().Foreground(styles.ColorBorder).Render(strings.Repeat("─", width))

	row := func(k, desc string) string {
		return helpRow(key(k), dim(desc))
	}

	var b strings.Builder

	// Navigation
	b.WriteString(h("Navigation") + "\n")
	b.WriteString(sep + "\n")
	b.WriteString(row("h / ←", "Colonne précédente") + "\n")
	b.WriteString(row("l / →", "Colonne suivante") + "\n")
	b.WriteString(row("j / ↓", "Tâche suivante") + "\n")
	b.WriteString(row("k / ↑", "Tâche précédente") + "\n")
	b.WriteString("\n")

	// Actions rapides
	b.WriteString(h("Actions rapides") + "\n")
	b.WriteString(sep + "\n")
	b.WriteString(row("n", "Nouvelle tâche dans la colonne active") + "\n")
	b.WriteString(row("e", "Éditer la tâche sélectionnée") + "\n")
	b.WriteString(row("space", "Avancer la tâche à la colonne suivante") + "\n")
	b.WriteString(row("dd", "Supprimer la tâche (confirmation requise)") + "\n")
	b.WriteString(row("enter", "Inspecter la tâche (checklist, détails)") + "\n")
	b.WriteString(row("q", "Quitter l'application") + "\n")
	b.WriteString("\n")

	// Commandes slash
	b.WriteString(h("Commandes  (/ ou :)") + "\n")
	b.WriteString(sep + "\n")
	b.WriteString(row("/add <titre>", "Nouvelle tâche") + "\n")
	b.WriteString(row("/add -q <titre>", "Nouvelle tâche sans formulaire") + "\n")
	b.WriteString(row("/add bug|feat|doc <titre>", "Nouvelle tâche avec type") + "\n")
	b.WriteString(row("/edit <id>", "Éditer une tâche par ID") + "\n")
	b.WriteString(row("/delete <id>", "Supprimer une tâche par ID") + "\n")
	b.WriteString(row("/move <id> <colonne>", "Déplacer une tâche vers une colonne") + "\n")
	b.WriteString(row("/sub-add <texte>", "Ajouter une sous-tâche à la tâche active") + "\n")
	b.WriteString(row("/project <nom>", "Changer de projet") + "\n")
	b.WriteString(row("/column-add <nom>", "Ajouter une colonne") + "\n")
	b.WriteString(row("/column-rename <id> <nom>", "Renommer une colonne") + "\n")
	b.WriteString(row("/column-delete <id>", "Supprimer une colonne") + "\n")
	b.WriteString(row("/column-left <id>", "Déplacer colonne vers la gauche") + "\n")
	b.WriteString(row("/column-right <id>", "Déplacer colonne vers la droite") + "\n")
	b.WriteString(row("/projects-dir <chemin>", "Changer le répertoire des projets") + "\n")
	b.WriteString(row("/help", "Afficher cette aide") + "\n")
	b.WriteString(row("/quit  /q", "Quitter") + "\n")
	b.WriteString("\n")

	// Formulaire de tâche
	b.WriteString(h("Formulaire de tâche  (n / /add / /edit)") + "\n")
	b.WriteString(sep + "\n")
	b.WriteString(row("Tab / Shift+Tab", "Naviguer entre les champs") + "\n")
	b.WriteString(row("◄ ►  (champ Type)", "Changer le type bug / feat / task / doc") + "\n")
	b.WriteString(row("Ctrl+S", "Sauvegarder") + "\n")
	b.WriteString(row("Esc", "Annuler") + "\n")
	b.WriteString("\n")

	// Inspecteur
	b.WriteString(h("Inspecteur de tâche  (enter)") + "\n")
	b.WriteString(sep + "\n")
	b.WriteString(row("j / k", "Naviguer dans la checklist") + "\n")
	b.WriteString(row("space", "Cocher / décocher un item") + "\n")
	b.WriteString(row("a", "Ajouter un item à la checklist") + "\n")
	b.WriteString(row("d", "Supprimer l'item sélectionné") + "\n")
	b.WriteString(row("e", "Éditer la tâche") + "\n")
	b.WriteString(row("Esc", "Fermer l'inspecteur") + "\n")
	b.WriteString("\n")

	// Stockage
	b.WriteString(h("Stockage") + "\n")
	b.WriteString(sep + "\n")
	b.WriteString(dim("Tâches  : ~/.kanban/<projet>/TASK-NNN.md") + "\n")
	b.WriteString(dim("Config  : ~/.kanban/config.yaml"))

	return b.String()
}

// helpRow aligne la touche sur une colonne fixe et la description après.
func helpRow(keyStr, desc string) string {
	const keyCol = 32
	kw := lipgloss.Width(keyStr)
	pad := keyCol - kw
	if pad < 1 {
		pad = 1
	}
	return keyStr + strings.Repeat(" ", pad) + desc
}
