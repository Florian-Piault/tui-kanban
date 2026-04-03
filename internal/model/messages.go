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

// Config columns
type ColumnAddMsg struct{ Name string }
type ColumnRenameMsg struct{ ID, NewName string }
type ColumnDeleteMsg struct{ ID string }
type ColumnMoveMsg struct {
	ID        string
	Direction int // -1 gauche, +1 droite
}

// Config globals
type ProjectsDirMsg struct{ Path string }

// Sauvegarde config
type configSavedMsg struct{}

// Inspect / checklists
type CloseInspectMsg struct{}
type openInspectMsg struct{ task storage.Task }
type checklistUpdatedMsg struct{ task storage.Task }
type checklistSavedMsg struct{ task storage.Task }
