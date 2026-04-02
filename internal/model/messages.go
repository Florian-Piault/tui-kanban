package model

import (
	"github.com/piflorian/tui-kanban/internal/command"
	"github.com/piflorian/tui-kanban/internal/storage"
)

// Navigation
type ColChangedMsg struct{ Index int }
type TaskSelectedMsg struct{ Task storage.Task }

// Chargement des tâches
type TasksLoadedMsg struct {
	ColID string
	Tasks []storage.Task
}

// CRUD tâches
type TaskCreatedMsg struct{ Task storage.Task }
type TaskUpdatedMsg struct{ Task storage.Task }
type TaskDeletedMsg struct{ ID string }
type TaskMovedMsg struct {
	ID    string
	ToCol string
}

// Projets
type ProjectChangedMsg struct{ Name string }
type ProjectsLoadedMsg struct{ Projects []string }
type TaskIDsLoadedMsg struct{ IDs []string }

// UI
type ErrMsg struct{ Err error }
type SuccessMsg struct{ Text string }
type ClearFlashMsg struct{}
type CloseCommandBarMsg struct{}
type CloseModalMsg struct{}
type OpenModalMsg struct {
	Task    storage.Task
	IsNew   bool
	ColID   string
}
type ConfirmDeleteMsg struct{ ID string }

// CommandParsedMsg est émis quand une slash command est parsée avec succès.
type CommandParsedMsg struct {
	Parsed command.ParsedCommand
}
