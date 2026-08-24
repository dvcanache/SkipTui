package views

import (
	"skiptui/internal/tui/styles"

	"github.com/charmbracelet/lipgloss"
)

// RenderLogs renders the diagnostic and session logs viewport.
func RenderLogs(logs []string, width int) string {
	var rows []string

	header := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Render("EVENT & CONNECTION LOGS")
	rows = append(rows, header, lipgloss.NewStyle().Foreground(styles.ColorBorder).Render("─"+repeatStr("─", width-4)))

	if len(logs) == 0 {
		rows = append(rows, lipgloss.NewStyle().Foreground(styles.ColorMuted).Padding(2, 2).Render("No log entries yet."))
	} else {
		// Display recent logs (up to 15 lines)
		start := 0
		if len(logs) > 15 {
			start = len(logs) - 15
		}
		for _, line := range logs[start:] {
			rows = append(rows, lipgloss.NewStyle().Foreground(styles.ColorFg).Render(line))
		}
	}

	return styles.BoxStyle.Width(width - 2).Render(
		lipgloss.JoinVertical(lipgloss.Left, rows...),
	)
}
