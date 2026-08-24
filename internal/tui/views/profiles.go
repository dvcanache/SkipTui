package views

import (
	"fmt"
	"skiptui/internal/config"
	"skiptui/internal/tui/styles"

	"github.com/charmbracelet/lipgloss"
)

// RenderProfiles renders the configured proxy and VPN profiles table.
func RenderProfiles(profiles []*config.Profile, selectedIdx int, width int) string {
	var rows []string

	// Table Header
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Width(20).Render("NAME"),
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Width(14).Render("PROTOCOL"),
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Width(28).Render("ENDPOINT"),
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Width(14).Render("AUTH"),
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Width(12).Render("LATENCY"),
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Width(10).Render("STATUS"),
	)
	rows = append(rows, header, lipgloss.NewStyle().Foreground(styles.ColorBorder).Render("─"+repeatStr("─", width-4)))

	if len(profiles) == 0 {
		emptyNotice := lipgloss.NewStyle().
			Foreground(styles.ColorMuted).
			Padding(2, 2).
			Render("No profiles found. Press [i] to import .ovpn / .conf or [a] to add a profile.")
		rows = append(rows, emptyNotice)
	} else {
		for i, p := range profiles {
			latencyStr := "--"
			statusBadge := lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("○ UNK")

			if p.LatencyMs > 0 {
				latencyStr = fmt.Sprintf("%d ms", p.LatencyMs)
				if p.LatencyMs < 100 {
					statusBadge = styles.BadgeSuccess.Render("● OK")
				} else if p.LatencyMs < 300 {
					statusBadge = styles.BadgeWarning.Render("● FAIR")
				} else {
					statusBadge = styles.BadgeDanger.Render("● SLOW")
				}
			} else if p.LatencyMs == -1 {
				latencyStr = "FAIL"
				statusBadge = styles.BadgeDanger.Render("✗ ERR")
			}

			authStr := "None"
			if p.Username != "" {
				authStr = "User/Pass"
			} else if p.Protocol == config.ProtocolWireGuard {
				authStr = "PubKey"
			} else if p.Protocol == config.ProtocolOpenVPN {
				authStr = "Cert/Pass"
			}

			rowContent := lipgloss.JoinHorizontal(
				lipgloss.Left,
				lipgloss.NewStyle().Width(20).Render(truncate(p.Name, 18)),
				lipgloss.NewStyle().Width(14).Render(string(p.Protocol)),
				lipgloss.NewStyle().Width(28).Render(truncate(p.Endpoint, 26)),
				lipgloss.NewStyle().Width(14).Render(authStr),
				lipgloss.NewStyle().Width(12).Render(latencyStr),
				lipgloss.NewStyle().Width(10).Render(statusBadge),
			)

			if i == selectedIdx {
				rows = append(rows, styles.SelectedRowStyle.Render(rowContent))
			} else {
				rows = append(rows, styles.NormalRowStyle.Render(rowContent))
			}
		}
	}

	profilesBox := styles.BoxStyle.Width(width - 2).Render(
		lipgloss.JoinVertical(lipgloss.Left, rows...),
	)

	// Detail Pane for selected profile
	var detailsBox string
	if len(profiles) > 0 && selectedIdx >= 0 && selectedIdx < len(profiles) {
		cur := profiles[selectedIdx]

		killSwitchStr := "Enabled (Fail-Closed)"
		if !cur.KillSwitch {
			killSwitchStr = "Disabled"
		}

		leftPane := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorBorder).
			Padding(0, 1).
			Width((width - 4) / 2).
			Render(lipgloss.JoinVertical(
				lipgloss.Left,
				fmt.Sprintf("Profile ID: %s", cur.ID),
				fmt.Sprintf("Protocol: %s", cur.Protocol),
				fmt.Sprintf("DNS Server: %s", cur.DNS),
			))

		rightPane := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorBorder).
			Padding(0, 1).
			Width((width - 4) / 2).
			Render(lipgloss.JoinVertical(
				lipgloss.Left,
				fmt.Sprintf("Endpoint: %s", cur.Endpoint),
				fmt.Sprintf("KillSwitch: %s", killSwitchStr),
				fmt.Sprintf("Last Tested: %s", cur.LastTested.Format("15:04:05")),
			))

		detailsBox = lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
	}

	return lipgloss.JoinVertical(lipgloss.Left, profilesBox, detailsBox)
}
