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
)

const reservedLines = 4 // header + séparateur + commandbar/statusbar + marge

type AppModel struct {
	state   AppState
	cfg     *config.Config
	storage *storage.Storage

	board      BoardModel
	commandBar CommandBarModel
	modal      ModalModel

	flash        string
	flashIsError bool
	flashTimer   int

	confirmID string // ID de tâche en attente de confirmation de suppression

	width  int
	height int
}

func New(cfg *config.Config, store *storage.Storage) AppModel {
	return AppModel{
		cfg:        cfg,
		storage:    store,
		board:      NewBoard(cfg),
		commandBar: NewCommandBar(),
		modal:      NewModal(),
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
		return m, tea.Batch(m.loadAllColumns(), m.loadContext())

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
	}
	return m, nil
}

type taskSavedMsg struct{ task storage.Task }
type taskDeletedOKMsg struct{ id string }
type taskMovedOKMsg struct{ id, toCol string }

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

	case StateBrowsing:
		switch msg.String() {
		case "/", ":":
			m.state = StateCommanding
			m.commandBar = NewCommandBar()
			m.commandBar.SetWidth(m.width)
			return m, m.loadContext()
		case "q", "ctrl+c":
			return m, tea.Quit
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
					return OpenModalMsg{Task: t, IsNew: false, ColID: t.Status}
				}
			}
		default:
			m.board.Update(msg)
		}
	}
	return m, nil
}

func (m AppModel) handleCommand(parsed command.ParsedCommand) (tea.Model, tea.Cmd) {
	m.state = StateBrowsing
	switch parsed.Name {
	case "quit":
		return m, tea.Quit

	case "add":
		colID := m.board.ActiveColumnID()
		title := ""
		if len(parsed.Args) > 0 {
			title = parsed.Args[0]
		}
		if parsed.Flags["q"] && title != "" {
			return m, func() tea.Msg {
				return TaskCreatedMsg{Task: storage.Task{Title: title, Status: colID}}
			}
		}
		return m, func() tea.Msg {
			return OpenModalMsg{
				Task:  storage.Task{Title: title, Status: colID},
				IsNew: true,
				ColID: colID,
			}
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

	case "help":
		m.setFlash("Commandes : /add /edit /delete /move /project /quit  •  h/l/j/k : navigation", false)
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

	return lipgloss.JoinVertical(lipgloss.Left, header, m.board.ViewAtHeight(boardH), bottom)
}

func (m AppModel) renderHeader() string {
	project := styles.StatusBarProjectStyle.Render("📋 " + m.cfg.CurrentProject)
	hint := styles.HelpStyle.Render("  /commande  •  h/l/j/k : navigation  •  q : quitter")
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
