package tui

import (
	"skiptui/internal/config"
	"skiptui/internal/session"
	"skiptui/internal/tui/modals"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type TabType int

const (
	TabSessions TabType = iota
	TabProfiles
	TabLogs
	TabSettings
)

// Model represents the top-level Bubble Tea Elm state.
type Model struct {
	ActiveTab    TabType
	Config       *config.Config
	Supervisor   *session.Supervisor
	Sessions     []*config.SessionInfo
	Profiles     []*config.Profile
	SelectedSess int
	SelectedProf int
	Logs         []string
	Width        int
	Height       int

	// Active Modals
	ShowLauncher bool
	Launcher     modals.LauncherModal

	ShowImporter bool
	Importer     modals.ImportWizardModal

	ShowExitDialog bool
	ShowHelp       bool

	StatusMsg string
}

// TickMsg triggers periodic refresh of session stats and background health.
type TickMsg time.Time

// LatencyResultMsg carries the outcome of profile latency tests.
type LatencyResultMsg struct {
	ProfileID string
	LatencyMs int64
	Error     error
}

func InitialModel(cfg *config.Config, sup *session.Supervisor) Model {
	sessions := sup.ListSessions()
	profiles := cfg.Profiles

	return Model{
		ActiveTab:      TabSessions,
		Config:         cfg,
		Supervisor:     sup,
		Sessions:       sessions,
		Profiles:       profiles,
		SelectedSess:   0,
		SelectedProf:   0,
		Logs:           []string{"SkipTUI initialized. Ready."},
		Width:          100,
		Height:         28,
		Launcher:       modals.NewLauncherModal(profiles),
		Importer:       modals.NewImportWizardModal(),
		ShowLauncher:   false,
		ShowImporter:   false,
		ShowExitDialog: false,
		ShowHelp:       false,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickEvery(1*time.Second),
	)
}

func tickEvery(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
