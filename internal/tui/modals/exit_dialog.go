package modals

import (
	"skiptui/internal/tui/styles"

	"github.com/charmbracelet/lipgloss"
)

// RenderExitDialog renders the Detach vs Terminate prompt.
func RenderExitDialog(activeSessionCount int, width int) string {
	dialogBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(styles.ColorWarning).
		Padding(1, 2).
		Width(64)

	title := styles.TitleStyle.Copy().Foreground(styles.ColorWarning).Render("Active Sessions Running")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		lipgloss.NewStyle().Foreground(styles.ColorFg).Render("You have active isolated sessions running in the background:"),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Render("[D] Detach & Keep Running in Background (Recommended)"),
		lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("    Sessions & tunnels remain active. Reconnect by reopening SkipTUI."),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorDanger).Render("[K] Terminate All Sessions & Clean Up Namespaces"),
		lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("    Stops all processes and destroys all temporary network sandboxes."),
		"",
		lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("[Esc] Cancel / Return to Dashboard"),
	)

	return lipgloss.Place(
		width,
		18,
		lipgloss.Center,
		lipgloss.Center,
		dialogBox.Render(content),
	)
}
