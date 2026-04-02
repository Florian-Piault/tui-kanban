package styles

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary   = lipgloss.Color("#7C3AED")
	ColorSecondary = lipgloss.Color("#A78BFA")
	ColorMuted     = lipgloss.Color("#6B7280")
	ColorBg        = lipgloss.Color("#1F2937")
	ColorBgActive  = lipgloss.Color("#374151")
	ColorText      = lipgloss.Color("#F9FAFB")
	ColorTextDim   = lipgloss.Color("#9CA3AF")
	ColorSuccess   = lipgloss.Color("#10B981")
	ColorError     = lipgloss.Color("#EF4444")
	ColorWarning   = lipgloss.Color("#F59E0B")
	ColorDue       = lipgloss.Color("#F59E0B")
	ColorBorder    = lipgloss.Color("#4B5563")
	ColorBorderAct = lipgloss.Color("#7C3AED")
)

var (
	ColumnStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	ColumnActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorderAct).
				Padding(0, 1)

	ColumnTitleStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Bold(true).
				Padding(0, 1)

	ColumnTitleActiveStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				Bold(true).
				Padding(0, 1)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1).
			MarginBottom(1)

	CardSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1).
				MarginBottom(1)

	DueStyle = lipgloss.NewStyle().
			Foreground(ColorDue).
			Italic(true)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim).
			Padding(0, 1)

	StatusBarProjectStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true).
			Padding(0, 1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 1)

	CommandBarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	SuggestionStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim).
			Padding(0, 1)

	SuggestionSelectedStyle = lipgloss.NewStyle().
				Background(ColorBgActive).
				Foreground(ColorSecondary).
				Padding(0, 1)

	ModalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	ModalTitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true).
			MarginBottom(1)

	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim).
			Width(12)
)
