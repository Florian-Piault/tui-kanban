package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/piflorian/tui-kanban/internal/command"
	"github.com/piflorian/tui-kanban/internal/config"
	"github.com/piflorian/tui-kanban/internal/storage"
	"github.com/piflorian/tui-kanban/internal/styles"
)

type AppState int

const (
	StateBrowsing   AppState = iota
	StateCommanding          // barre de commandes ouverte
	StateEditing             // modal formulaire ouvert
	StateConfirming          // confirmation de suppression
	StateInspecting          // vue détaillée d'une tâche + checklists
)

const reservedLines = 2 // header(1) + statusbar(1)

type AppModel struct {
	state   AppState
	cfg     *config.Config
	cfgPath string
	storage *storage.Storage

	board      BoardModel
	commandBar CommandBarModel
	modal      ModalModel
	inspect    InspectModel

	flash        string
	flashIsError bool
	flashTimer   int

	confirmID string // ID de tâche en attente de confirmation de suppression
	pendingD  bool   // true si un "d" a été tapé, attend un second "d" pour supprimer

	width  int
	height int
}

func New(cfg *config.Config, store *storage.Storage, cfgPath string) AppModel {
	return AppModel{
		cfg:        cfg,
		cfgPath:    cfgPath,
		storage:    store,
		board:      NewBoard(cfg),
		commandBar: NewCommandBar(),
		modal:      NewModal(),
		inspect:    NewInspectModel(),
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.loadAllColumns()
}

// --- Init / chargement ---

func (m AppModel) loadAllColumns() tea.Cmd {
	var cmds []tea.Cmd
	for _, col := range m.cfg.Columns {
		colID := col.ID
		project := m.cfg.CurrentProject
		cmds = append(cmds, func() tea.Msg {
			tasks, err := m.storage.LoadByStatus(project, colID)
			if err != nil {
				return ErrMsg{Err: err}
			}
			return TasksLoadedMsg{ColID: colID, Tasks: tasks}
		})
	}
	return tea.Batch(cmds...)
}

func (m AppModel) loadColumn(colID string) tea.Cmd {
	project := m.cfg.CurrentProject
	return func() tea.Msg {
		tasks, err := m.storage.LoadByStatus(project, colID)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return TasksLoadedMsg{ColID: colID, Tasks: tasks}
	}
}

func (m AppModel) loadContext() tea.Cmd {
	project := m.cfg.CurrentProject
	store := m.storage
	return func() tea.Msg {
		tasks, err := store.LoadAll(project)
		if err != nil {
			return ErrMsg{Err: err}
		}
		ids := make([]string, len(tasks))
		titles := make(map[string]string, len(tasks))
		for i, t := range tasks {
			ids[i] = t.ID
			titles[t.ID] = t.Title
		}
		projects, err := store.ListProjects()
		if err != nil {
			projects = nil
		}
		return contextLoadedMsg{taskIDs: ids, taskTitles: titles, projects: projects}
	}
}

type contextLoadedMsg struct {
	taskIDs    []string
	taskTitles map[string]string
	projects   []string
}

// --- Update ---

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		boardH := msg.Height - reservedLines
		if boardH < 5 {
			boardH = 5
		}
		m.board.SetSize(msg.Width, boardH)
		m.commandBar.SetWidth(msg.Width)
		m.modal.Width = msg.Width
		m.modal.applyInputWidths()
		m.inspect.Width = msg.Width
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case TasksLoadedMsg:
		m.board.SetTasks(msg.ColID, msg.Tasks)
		return m, nil

	case contextLoadedMsg:
		colIDs := make([]string, len(m.cfg.Columns))
		for i, c := range m.cfg.Columns {
			colIDs[i] = c.ID
		}
		ctx := command.CompletionContext{
			TaskIDs:    msg.taskIDs,
			TaskTitles: msg.taskTitles,
			ColumnIDs:  colIDs,
			ProjectIDs: msg.projects,
		}
		m.commandBar.SetContext(ctx)
		m.commandBar.refreshSuggestions()
		return m, nil

	case TaskCreatedMsg:
		project := m.cfg.CurrentProject
		task := msg.Task
		store := m.storage
		return m, tea.Batch(
			func() tea.Msg {
				saved, err := store.SaveTask(project, task)
				if err != nil {
					return ErrMsg{Err: err}
				}
				return taskSavedMsg{task: saved}
			},
		)

	case taskSavedMsg:
		m.state = StateBrowsing
		m.setFlash("✓ Tâche sauvegardée", false)
		return m, tea.Batch(m.loadAllColumns(), m.loadContext())

	case openInspectMsg:
		m.state = StateInspecting
		m.inspect.Width = m.width
		m.inspect.Open(msg.task)
		return m, nil

	case CloseInspectMsg:
		m.state = StateBrowsing
		return m, tea.ClearScreen

	case checklistUpdatedMsg:
		project := m.cfg.CurrentProject
		task := msg.task
		store := m.storage
		return m, func() tea.Msg {
			saved, err := store.SaveTask(project, task)
			if err != nil {
				return ErrMsg{Err: err}
			}
			return checklistSavedMsg{task: saved}
		}

	case checklistSavedMsg:
		m.inspect.task = msg.task
		return m, m.loadAllColumns()

	case TaskUpdatedMsg:
		project := m.cfg.CurrentProject
		task := msg.Task
		store := m.storage
		return m, func() tea.Msg {
			saved, err := store.SaveTask(project, task)
			if err != nil {
				return ErrMsg{Err: err}
			}
			return taskSavedMsg{task: saved}
		}

	case ConfirmDeleteMsg:
		m.confirmID = msg.ID
		m.state = StateConfirming
		return m, nil

	case TaskDeletedMsg:
		project := m.cfg.CurrentProject
		id := msg.ID
		store := m.storage
		return m, func() tea.Msg {
			if err := store.DeleteTask(project, id); err != nil {
				return ErrMsg{Err: err}
			}
			return taskDeletedOKMsg{id: id}
		}

	case taskDeletedOKMsg:
		m.state = StateBrowsing
		m.setFlash(fmt.Sprintf("✓ %s supprimée", msg.id), false)
		return m, tea.Batch(m.loadAllColumns(), m.loadContext())

	case TaskMovedMsg:
		project := m.cfg.CurrentProject
		id, toCol := msg.ID, msg.ToCol
		store := m.storage
		return m, func() tea.Msg {
			_, err := store.MoveTask(project, id, toCol)
			if err != nil {
				return ErrMsg{Err: err}
			}
			return taskMovedOKMsg{id: id, toCol: toCol}
		}

	case taskMovedOKMsg:
		m.setFlash(fmt.Sprintf("✓ %s → %s", msg.id, msg.toCol), false)
		return m, m.loadAllColumns()

	case ProjectChangedMsg:
		m.cfg.CurrentProject = msg.Name
		m.board = NewBoard(m.cfg)
		m.board.SetSize(m.width, m.height-reservedLines)
		m.setFlash("Projet : "+msg.Name, false)
		return m, tea.Batch(m.saveConfig(), m.loadAllColumns(), m.loadContext())

	case ColumnAddMsg:
		col, err := m.cfg.AddColumn(msg.Name)
		if err != nil {
			m.setFlash("✗ "+err.Error(), true)
			return m, nil
		}
		m.rebuildBoard()
		m.setFlash(fmt.Sprintf("✓ Colonne %q ajoutée (ID : %s)", col.Name, col.ID), false)
		return m, tea.Batch(m.saveConfig(), m.loadAllColumns(), m.loadContext())

	case ColumnRenameMsg:
		if err := m.cfg.RenameColumn(msg.ID, msg.NewName); err != nil {
			m.setFlash("✗ "+err.Error(), true)
			return m, nil
		}
		m.rebuildBoard()
		m.setFlash(fmt.Sprintf("✓ Colonne %q renommée en %q", msg.ID, msg.NewName), false)
		return m, tea.Batch(m.saveConfig(), m.loadAllColumns(), m.loadContext())

	case ColumnDeleteMsg:
		if err := m.cfg.DeleteColumn(msg.ID); err != nil {
			m.setFlash("✗ "+err.Error(), true)
			return m, nil
		}
		m.rebuildBoard()
		m.setFlash(fmt.Sprintf("✓ Colonne %q supprimée", msg.ID), false)
		return m, tea.Batch(m.saveConfig(), m.loadContext())

	case ColumnMoveMsg:
		var err error
		if msg.Direction < 0 {
			err = m.cfg.MoveColumnLeft(msg.ID)
		} else {
			err = m.cfg.MoveColumnRight(msg.ID)
		}
		if err != nil {
			m.setFlash("✗ "+err.Error(), true)
			return m, nil
		}
		m.rebuildBoard()
		dir := "droite"
		if msg.Direction < 0 {
			dir = "gauche"
		}
		m.setFlash(fmt.Sprintf("✓ Colonne %q déplacée vers la %s", msg.ID, dir), false)
		return m, tea.Batch(m.saveConfig(), m.loadAllColumns(), m.loadContext())

	case ProjectsDirMsg:
		m.cfg.ProjectsDir = msg.Path
		m.storage = storage.New(msg.Path)
		m.board = NewBoard(m.cfg)
		m.board.SetSize(m.width, m.height-reservedLines)
		m.setFlash("✓ Répertoire : "+msg.Path, false)
		return m, tea.Batch(m.saveConfig(), m.loadAllColumns(), m.loadContext())

	case configSavedMsg:
		return m, nil

	case CommandParsedMsg:
		return m.handleCommand(msg.Parsed)

	case OpenModalMsg:
		m.state = StateEditing
		m.modal.Open(msg.Task, msg.IsNew, msg.ColID)
		return m, nil

	case CloseCommandBarMsg:
		m.state = StateBrowsing
		return m, nil

	case CloseModalMsg:
		m.state = StateBrowsing
		return m, nil

	case ErrMsg:
		m.setFlash("✗ "+msg.Err.Error(), true)
		m.state = StateBrowsing
		return m, nil

	case SuccessMsg:
		m.setFlash(msg.Text, false)
		return m, nil

	case ClearFlashMsg:
		m.flash = ""
		return m, nil
	}

	// Délégation selon l'état
	switch m.state {
	case StateCommanding:
		var cmd tea.Cmd
		m.commandBar, cmd = m.commandBar.Update(msg)
		return m, cmd
	case StateEditing:
		var cmd tea.Cmd
		m.modal, cmd = m.modal.Update(msg)
		return m, cmd
	case StateInspecting:
		var cmd tea.Cmd
		m.inspect, cmd = m.inspect.Update(msg)
		return m, cmd
	}
	return m, nil
}

type taskSavedMsg struct{ task storage.Task }
type taskDeletedOKMsg struct{ id string }
type taskMovedOKMsg struct{ id, toCol string }

// saveConfig persiste la config sur disque de façon asynchrone.
func (m AppModel) saveConfig() tea.Cmd {
	cfg := m.cfg
	path := m.cfgPath
	return func() tea.Msg {
		if err := config.Save(cfg, path); err != nil {
			return ErrMsg{Err: err}
		}
		return configSavedMsg{}
	}
}

// rebuildBoard reconstruit le board depuis la config en préservant la colonne active.
func (m *AppModel) rebuildBoard() {
	activeID := m.board.ActiveColumnID()
	m.board = NewBoard(m.cfg)
	m.board.SetSize(m.width, m.height-reservedLines)
	for i, col := range m.board.Columns {
		if col.ID == activeID {
			m.board.Columns[m.board.ActiveCol].IsActive = false
			m.board.ActiveCol = i
			m.board.Columns[i].IsActive = true
			break
		}
	}
}

func (m *AppModel) setFlash(text string, isError bool) {
	m.flash = text
	m.flashIsError = isError
	// Auto-clear après 3 secondes
}

func (m AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateCommanding:
		var cmd tea.Cmd
		m.commandBar, cmd = m.commandBar.Update(msg)
		return m, cmd

	case StateEditing:
		var cmd tea.Cmd
		m.modal, cmd = m.modal.Update(msg)
		return m, cmd

	case StateConfirming:
		switch msg.String() {
		case "y", "Y", "o", "O":
			id := m.confirmID
			m.confirmID = ""
			m.state = StateBrowsing
			return m, func() tea.Msg { return TaskDeletedMsg{ID: id} }
		default:
			m.confirmID = ""
			m.state = StateBrowsing
			m.setFlash("Suppression annulée", false)
		}
		return m, nil

	case StateInspecting:
		if msg.String() == "/" || msg.String() == ":" {
			m.state = StateCommanding
			m.commandBar = NewCommandBar()
			m.commandBar.SetWidth(m.width)
			return m, m.loadContext()
		}
		var cmd tea.Cmd
		m.inspect, cmd = m.inspect.Update(msg)
		return m, cmd

	case StateBrowsing:
		// Réinitialise la séquence "dd" sur toute touche autre que "d"
		if msg.String() != "d" {
			m.pendingD = false
		}

		switch msg.String() {
		case "/", ":":
			m.state = StateCommanding
			m.commandBar = NewCommandBar()
			m.commandBar.SetWidth(m.width)
			return m, m.loadContext()
		case "q", "ctrl+c":
			return m, tea.Quit
		case "n":
			colID := m.board.ActiveColumnID()
			task := storage.Task{Status: colID, Type: storage.TypeTask}
			return m, func() tea.Msg {
				return OpenModalMsg{Task: task, IsNew: true, ColID: colID}
			}
		case "e":
			if task, ok := m.board.SelectedTask(); ok {
				project := m.cfg.CurrentProject
				store := m.storage
				id := task.ID
				return m, func() tea.Msg {
					t, err := store.GetTask(project, id)
					if err != nil {
						return ErrMsg{Err: err}
					}
					return OpenModalMsg{Task: t, IsNew: false, ColID: t.Status}
				}
			}
		case " ":
			if task, ok := m.board.SelectedTask(); ok {
				nextCol := m.nextColumnID(task.Status)
				if nextCol == "" {
					m.setFlash("Déjà dans la dernière colonne", false)
					return m, nil
				}
				id := task.ID
				return m, func() tea.Msg { return TaskMovedMsg{ID: id, ToCol: nextCol} }
			}
		case "d":
			if m.pendingD {
				m.pendingD = false
				if task, ok := m.board.SelectedTask(); ok {
					id := task.ID
					return m, func() tea.Msg { return ConfirmDeleteMsg{ID: id} }
				}
			} else {
				m.pendingD = true
			}
			return m, nil
		case "enter":
			if task, ok := m.board.SelectedTask(); ok {
				project := m.cfg.CurrentProject
				store := m.storage
				id := task.ID
				return m, func() tea.Msg {
					t, err := store.GetTask(project, id)
					if err != nil {
						return ErrMsg{Err: err}
					}
					return openInspectMsg{task: t}
				}
			}
		default:
			m.board.Update(msg)
		}
	}
	return m, nil
}

// nextColumnID retourne l'ID de la colonne suivant colID, ou "" si c'est la dernière.
func (m AppModel) nextColumnID(colID string) string {
	cols := m.cfg.Columns
	for i, c := range cols {
		if c.ID == colID && i+1 < len(cols) {
			return cols[i+1].ID
		}
	}
	return ""
}

func (m AppModel) handleCommand(parsed command.ParsedCommand) (tea.Model, tea.Cmd) {
	m.state = StateBrowsing
	switch parsed.Name {
	case "quit":
		return m, tea.Quit

	case "add":
		colID := m.board.ActiveColumnID()
		rawTitle := ""
		if len(parsed.Args) > 0 {
			rawTitle = parsed.Args[0]
		}
		// Détection du type en préfixe optionnel : /add bug Titre / /add feat Titre
		taskType := storage.TypeTask
		if rawTitle != "" {
			words := strings.SplitN(rawTitle, " ", 2)
			if storage.IsValidType(words[0]) {
				taskType = storage.NormalizeType(words[0])
				if len(words) > 1 {
					rawTitle = words[1]
				} else {
					rawTitle = ""
				}
			}
		}
		task := storage.Task{Title: rawTitle, Status: colID, Type: taskType}
		if parsed.Flags["q"] && rawTitle != "" {
			return m, func() tea.Msg {
				return TaskCreatedMsg{Task: task}
			}
		}
		return m, func() tea.Msg {
			return OpenModalMsg{Task: task, IsNew: true, ColID: colID}
		}

	case "edit":
		if len(parsed.Args) == 0 {
			return m, func() tea.Msg { return ErrMsg{Err: fmt.Errorf("ID requis")} }
		}
		id := parsed.Args[0]
		project := m.cfg.CurrentProject
		store := m.storage
		return m, func() tea.Msg {
			task, err := store.GetTask(project, id)
			if err != nil {
				return ErrMsg{Err: err}
			}
			return OpenModalMsg{Task: task, IsNew: false, ColID: task.Status}
		}

	case "delete":
		if len(parsed.Args) == 0 {
			return m, func() tea.Msg { return ErrMsg{Err: fmt.Errorf("ID requis")} }
		}
		return m, func() tea.Msg { return ConfirmDeleteMsg{ID: parsed.Args[0]} }

	case "move":
		if len(parsed.Args) < 2 {
			return m, func() tea.Msg { return ErrMsg{Err: fmt.Errorf("usage : /move <id> <colonne>")} }
		}
		return m, func() tea.Msg { return TaskMovedMsg{ID: parsed.Args[0], ToCol: parsed.Args[1]} }

	case "project":
		if len(parsed.Args) == 0 {
			return m, func() tea.Msg { return ErrMsg{Err: fmt.Errorf("nom de projet requis")} }
		}
		return m, func() tea.Msg { return ProjectChangedMsg{Name: parsed.Args[0]} }

	case "column-add":
		if len(parsed.Args) == 0 {
			return m, func() tea.Msg { return ErrMsg{Err: fmt.Errorf("nom de colonne requis")} }
		}
		name := parsed.Args[0]
		return m, func() tea.Msg { return ColumnAddMsg{Name: name} }

	case "column-rename":
		if len(parsed.Args) < 2 {
			return m, func() tea.Msg { return ErrMsg{Err: fmt.Errorf("usage : /column-rename <id> <nouveau-nom>")} }
		}
		id, newName := parsed.Args[0], parsed.Args[1]
		return m, func() tea.Msg { return ColumnRenameMsg{ID: id, NewName: newName} }

	case "column-delete":
		if len(parsed.Args) == 0 {
			return m, func() tea.Msg { return ErrMsg{Err: fmt.Errorf("ID de colonne requis")} }
		}
		id := parsed.Args[0]
		return m, func() tea.Msg { return ColumnDeleteMsg{ID: id} }

	case "column-left":
		if len(parsed.Args) == 0 {
			return m, func() tea.Msg { return ErrMsg{Err: fmt.Errorf("ID de colonne requis")} }
		}
		id := parsed.Args[0]
		return m, func() tea.Msg { return ColumnMoveMsg{ID: id, Direction: -1} }

	case "column-right":
		if len(parsed.Args) == 0 {
			return m, func() tea.Msg { return ErrMsg{Err: fmt.Errorf("ID de colonne requis")} }
		}
		id := parsed.Args[0]
		return m, func() tea.Msg { return ColumnMoveMsg{ID: id, Direction: +1} }

	case "projects-dir":
		if len(parsed.Args) == 0 {
			return m, func() tea.Msg { return ErrMsg{Err: fmt.Errorf("chemin requis")} }
		}
		path := parsed.Args[0]
		return m, func() tea.Msg { return ProjectsDirMsg{Path: path} }

	case "sub-add":
		if len(parsed.Args) == 0 {
			return m, func() tea.Msg { return ErrMsg{Err: fmt.Errorf("texte de sous-tâche requis")} }
		}
		text := parsed.Args[0]
		task, ok := m.board.SelectedTask()
		if !ok {
			return m, func() tea.Msg { return ErrMsg{Err: fmt.Errorf("aucune tâche sélectionnée")} }
		}
		taskID := task.ID
		project := m.cfg.CurrentProject
		store := m.storage
		return m, func() tea.Msg {
			t, err := store.GetTask(project, taskID)
			if err != nil {
				return ErrMsg{Err: err}
			}
			t.Checklist = append(t.Checklist, storage.ChecklistItem{Text: text})
			saved, err := store.SaveTask(project, t)
			if err != nil {
				return ErrMsg{Err: err}
			}
			return taskSavedMsg{task: saved}
		}

	case "help":
		m.setFlash("Commandes : /add /edit /delete /move /sub-add /project /column-add /column-rename /column-delete /column-left /column-right /projects-dir /quit", false)
	}
	return m, nil
}

// --- View ---

func (m AppModel) View() string {
	header := m.renderHeader()

	// Zone basse rendue en premier pour mesurer sa hauteur
	var bottom string
	switch m.state {
	case StateCommanding:
		bottom = m.commandBar.View()
	case StateEditing:
		bottom = m.modal.View()
	case StateInspecting:
		bottom = m.inspect.View()
	case StateConfirming:
		bottom = styles.ErrorStyle.Render(fmt.Sprintf("Supprimer %s ? (y/n)", m.confirmID))
	default:
		bottom = m.renderStatusBar()
	}

	// Hauteur disponible pour le board = total - header(1) - bottom
	bottomLines := strings.Count(bottom, "\n") + 1
	boardH := m.height - 1 - bottomLines
	if boardH < 1 {
		boardH = 1
	}

	view := lipgloss.JoinVertical(lipgloss.Left, header, m.board.ViewAtHeight(boardH), bottom)
	return m.padToWidth(view)
}

// padToWidth force chaque ligne du rendu à exactement m.width caractères visibles.
// Sans ce padding, les artefacts de l'ancien contenu (plus large) restent à l'écran
// quand on passe d'une vue large (inspect, modal) à une vue plus étroite (statusbar).
func (m AppModel) padToWidth(view string) string {
	if m.width <= 0 {
		return view
	}
	lines := strings.Split(view, "\n")
	for i, line := range lines {
		w := lipgloss.Width(line)
		if w < m.width {
			lines[i] = line + strings.Repeat(" ", m.width-w)
		}
	}
	return strings.Join(lines, "\n")
}

func (m AppModel) renderHeader() string {
	project := styles.StatusBarProjectStyle.Render(m.cfg.CurrentProject)
	hint := styles.HelpStyle.Render("  /commande  •  hjkl : navigation  •  n : ajouter  •  e : éditer  •  space : avancer  •  dd : supprimer  •  enter : inspecter  •  q : quitter")
	return lipgloss.JoinHorizontal(lipgloss.Top, project, hint)
}

func (m AppModel) renderStatusBar() string {
	var parts []string

	if m.flash != "" {
		if m.flashIsError {
			parts = append(parts, styles.ErrorStyle.Render(m.flash))
		} else {
			parts = append(parts, styles.SuccessStyle.Render(m.flash))
		}
	} else {
		col := m.board.ActiveColumn()
		if col != nil {
			info := fmt.Sprintf("Colonne : %s (%d tâches)", col.Name, len(col.Tasks))
			parts = append(parts, styles.StatusBarStyle.Render(info))
		}
		parts = append(parts, styles.HelpStyle.Render(" • / pour les commandes"))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}
