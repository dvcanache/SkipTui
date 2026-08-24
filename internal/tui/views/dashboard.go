package views

import (
	"fmt"
	"skiptui/internal/config"
	"skiptui/internal/tui/styles"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// RenderDashboard renders the active isolated sessions table and detailed status panel.
func RenderDashboard(sessions []*config.SessionInfo, selectedIdx int, width int) string {
	var rows []string

	// Table Header
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Width(12).Render("ID"),
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Width(24).Render("COMMAND"),
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Width(18).Render("PROFILE"),
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Width(10).Render("PID"),
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Width(12).Render("UPTIME"),
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Width(14).Render("RX/TX"),
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Width(10).Render("STATUS"),
	)
	rows = append(rows, header, lipgloss.NewStyle().Foreground(styles.ColorBorder).Render("─"+repeatStr("─", width-4)))

	if len(sessions) == 0 {
		emptyNotice := lipgloss.NewStyle().
			Foreground(styles.ColorMuted).
			Padding(2, 2).
			Render("No active isolated sessions. Press [l] to launch an app or [t] to spawn an isolated terminal.")
		rows = append(rows, emptyNotice)
	} else {
		for i, s := range sessions {
			uptime := "00:00:00"
			if !s.StartTime.IsZero() {
				dur := time.Since(s.StartTime)
				uptime = fmt.Sprintf("%02d:%02d:%02d", int(dur.Hours()), int(dur.Minutes())%60, int(dur.Seconds())%60)
			}

			rxTx := formatBytes(s.BytesRX) + " / " + formatBytes(s.BytesTX)
			statusBadge := styles.BadgeSuccess.Render("● RUN")
			if s.Status == "stopped" {
				statusBadge = styles.BadgeDanger.Render("○ STOP")
			}

			pidStr := fmt.Sprintf("%d", s.PID)
			if s.PID == 0 {
				pidStr = "External"
			}

			rowContent := lipgloss.JoinHorizontal(
				lipgloss.Left,
				lipgloss.NewStyle().Width(12).Render(s.ID),
				lipgloss.NewStyle().Width(24).Render(truncate(s.Command, 22)),
				lipgloss.NewStyle().Width(18).Render(truncate(s.ProfileName, 16)),
				lipgloss.NewStyle().Width(10).Render(pidStr),
				lipgloss.NewStyle().Width(12).Render(uptime),
				lipgloss.NewStyle().Width(14).Render(rxTx),
				lipgloss.NewStyle().Width(10).Render(statusBadge),
			)

			if i == selectedIdx {
				rows = append(rows, styles.SelectedRowStyle.Render(rowContent))
			} else {
				rows = append(rows, styles.NormalRowStyle.Render(rowContent))
			}
		}
	}

	sessionsBox := styles.BoxStyle.Width(width - 2).Render(
		lipgloss.JoinVertical(lipgloss.Left, rows...),
	)

	// Detail Pane for selected session
	var detailsBox string
	if len(sessions) > 0 && selectedIdx >= 0 && selectedIdx < len(sessions) {
		cur := sessions[selectedIdx]
		leftPane := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorBorder).
			Padding(0, 1).
			Width((width - 4) / 2).
			Render(lipgloss.JoinVertical(
				lipgloss.Left,
				fmt.Sprintf("Namespace: %s", cur.Namespace),
				fmt.Sprintf("DNS Resolver: %s", cur.DNS),
				fmt.Sprintf("Started At: %s", cur.StartTime.Format("15:04:05")),
			))

		rightPane := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorBorder).
			Padding(0, 1).
			Width((width - 4) / 2).
			Render(lipgloss.JoinVertical(
				lipgloss.Left,
				fmt.Sprintf("Protocol: %s", cur.Protocol),
				fmt.Sprintf("KillSwitch: Fail-Closed Enforced"),
				fmt.Sprintf("Log File: %s", cur.LogFilePath),
			))

		detailsBox = lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sessionsBox, detailsBox)
}

func formatBytes(b uint64) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	} else if b < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
}

func truncate(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen-3] + "..."
}

func repeatStr(s string, n int) string {
	if n <= 0 {
		return ""
	}
	res := ""
	for i := 0; i < n; i++ {
		res += s
	}
	return res
}
