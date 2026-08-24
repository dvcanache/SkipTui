package modals

import (
	"fmt"
	"skiptui/internal/config"
	"skiptui/internal/tui/styles"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// LauncherModal represents the interactive state of the Quick Launcher.
type LauncherModal struct {
	CommandInput textinput.Model
	Profiles     []*config.Profile
	SelectedProf int
	InTerminal   bool
	Active       bool
}

func NewLauncherModal(profiles []*config.Profile) LauncherModal {
	ti := textinput.New()
	ti.Placeholder = "zsh (or firefox, curl api.ipify.org, etc.)"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 48

	return LauncherModal{
		CommandInput: ti,
		Profiles:     profiles,
		SelectedProf: 0,
		InTerminal:   true,
		Active:       false,
	}
}

func (m *LauncherModal) Render(width int) string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(1, 2).
		Width(64)

	title := styles.TitleStyle.Render("Launch Isolated Process / Terminal")

	var profOptions []string
	for i, p := range m.Profiles {
		indicator := "( )"
		style := lipgloss.NewStyle().Foreground(styles.ColorFg)
		if i == m.SelectedProf {
			indicator = "(•)"
			style = lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary)
		}
		profOptions = append(profOptions, style.Render(fmt.Sprintf("  %s %s [%s]", indicator, p.Name, p.Protocol)))
	}

	termCheckbox := "[ ]"
	if m.InTerminal {
		termCheckbox = "[X]"
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSecondary).Render("Command to Execute:"),
		m.CommandInput.View(),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSecondary).Render("Select Network Profile: (↑ / ↓ to select)"),
		lipgloss.JoinVertical(lipgloss.Left, profOptions...),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSecondary).Render("Execution Target: (Press [Tab] to toggle)"),
		fmt.Sprintf("  %s Launch in External Terminal Window (Kitty/Alacritty/Tmux)", termCheckbox),
		"",
		lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("[Enter] Launch Sandbox       [Esc] Cancel"),
	)

	return lipgloss.Place(
		width,
		22,
		lipgloss.Center,
		lipgloss.Center,
		box.Render(content),
	)
}
