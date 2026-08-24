package modals

import (
	"fmt"
	"os"
	"path/filepath"
	"skiptui/internal/config"
	"skiptui/internal/tui/styles"

	"github.com/charmbracelet/lipgloss"
)

// TerminalPickerModal represents the quick profile selection menu when pressing [t].
type TerminalPickerModal struct {
	Profiles    []*config.Profile
	SelectedIdx int
	Shell       string
	Active      bool
}

func NewTerminalPickerModal(profiles []*config.Profile) TerminalPickerModal {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "zsh"
	} else {
		shell = filepath.Base(shell)
	}

	return TerminalPickerModal{
		Profiles:    profiles,
		SelectedIdx: 0,
		Shell:       shell,
		Active:      false,
	}
}

func (m *TerminalPickerModal) Next() {
	if len(m.Profiles) > 0 {
		m.SelectedIdx = (m.SelectedIdx + 1) % len(m.Profiles)
	}
}

func (m *TerminalPickerModal) Prev() {
	if len(m.Profiles) > 0 {
		if m.SelectedIdx == 0 {
			m.SelectedIdx = len(m.Profiles) - 1
		} else {
			m.SelectedIdx--
		}
	}
}

func (m *TerminalPickerModal) GetSelectedProfile() *config.Profile {
	if len(m.Profiles) > 0 && m.SelectedIdx >= 0 && m.SelectedIdx < len(m.Profiles) {
		return m.Profiles[m.SelectedIdx]
	}
	return nil
}

func (m *TerminalPickerModal) Render(width int) string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(1, 2).
		Width(72)

	title := styles.TitleStyle.Render("⚡ Select Profile for Isolated Terminal")

	var rows []string
	if len(m.Profiles) == 0 {
		rows = append(rows, lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("No profiles found. Press [a] to add one or [i] to import .ovpn"))
	} else {
		for i, p := range m.Profiles {
			cursor := "  "
			style := lipgloss.NewStyle().Foreground(styles.ColorFg)
			if i == m.SelectedIdx {
				cursor = "▶ "
				style = lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Background(styles.ColorBorder)
			}

			latencyStr := "--"
			if p.LatencyMs > 0 {
				latencyStr = fmt.Sprintf("%dms", p.LatencyMs)
			}

			protoBadge := fmt.Sprintf("[%s]", string(p.Protocol))
			line := fmt.Sprintf("%s[%d] %-20s %-12s %-24s %6s",
				cursor,
				i+1,
				truncateStr(p.Name, 18),
				protoBadge,
				truncateStr(p.Endpoint, 22),
				latencyStr,
			)

			rows = append(rows, style.Render(line))
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		lipgloss.NewStyle().Foreground(styles.ColorSecondary).Render(fmt.Sprintf("Spawning interactive '%s' terminal shell in isolated network:", m.Shell)),
		"",
		lipgloss.JoinVertical(lipgloss.Left, rows...),
		"",
		lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("[↑ / ↓ / j / k] Select Profile    [Enter] Launch Terminal    [Esc] Cancel"),
	)

	return lipgloss.Place(
		width,
		22,
		lipgloss.Center,
		lipgloss.Center,
		box.Render(content),
	)
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
