package tui

import (
	"fmt"
	"skiptui/internal/config"
	"skiptui/internal/session"

	tea "github.com/charmbracelet/bubbletea"
)

// Start launches the interactive Bubble Tea terminal interface.
func Start(cfg *config.Config, sup *session.Supervisor) error {
	p := tea.NewProgram(
		InitialModel(cfg, sup),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}
