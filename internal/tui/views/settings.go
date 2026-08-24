package views

import (
	"fmt"
	"os"
	"os/exec"
	"skiptui/internal/config"
	"skiptui/internal/tui/styles"

	"github.com/charmbracelet/lipgloss"
)

// RenderSettings displays system capabilities and application preferences.
func RenderSettings(cfg *config.Config, width int) string {
	var rows []string

	header := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Render("SYSTEM CAPABILITIES & SETTINGS")
	rows = append(rows, header, lipgloss.NewStyle().Foreground(styles.ColorBorder).Render("─"+repeatStr("─", width-4)))

	// Check capability status
	rootStatus := styles.BadgeDanger.Render("✗ Non-Root (Standard User)")
	if os.Geteuid() == 0 {
		rootStatus = styles.BadgeSuccess.Render("✓ Root / Elevated")
	}

	ipStatus := styles.BadgeDanger.Render("✗ Missing")
	if _, err := exec.LookPath("ip"); err == nil {
		ipStatus = styles.BadgeSuccess.Render("✓ Installed")
	}

	ovpnStatus := styles.BadgeDanger.Render("✗ Not Found (Install openvpn for .ovpn)")
	if _, err := exec.LookPath("openvpn"); err == nil {
		ovpnStatus = styles.BadgeSuccess.Render("✓ Installed")
	}

	wgStatus := styles.BadgeWarning.Render("○ Kernel Module / Go Engine")

	section := func(title string, items ...string) string {
		var content []string
		content = append(content, lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSecondary).Render(title))
		for _, item := range items {
			content = append(content, "  "+item)
		}
		return lipgloss.JoinVertical(lipgloss.Left, content...)
	}

	capsSection := section(
		"Host System Environment:",
		fmt.Sprintf("Privilege State:       %s", rootStatus),
		fmt.Sprintf("Linux `ip` Utility:    %s", ipStatus),
		fmt.Sprintf("OpenVPN Binary:        %s", ovpnStatus),
		fmt.Sprintf("WireGuard Support:     %s", wgStatus),
	)

	pathsSection := section(
		"XDG Directory Paths:",
		fmt.Sprintf("Config Directory:      %s", config.GetConfigDir()),
		fmt.Sprintf("Profiles Directory:    %s", config.GetProfilesDir()),
		fmt.Sprintf("Runtime Socket:        %s", config.GetSocketPath()),
	)

	configSection := section(
		"Application Preferences:",
		fmt.Sprintf("Default Profile:       %s", cfg.Settings.DefaultProfile),
		fmt.Sprintf("Preferred Terminal:    %s", cfg.Settings.PreferredTerm),
		fmt.Sprintf("Fallback DNS:          %s", cfg.Settings.DNSFallback),
		fmt.Sprintf("Theme:                 %s", cfg.Settings.Theme),
	)

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		capsSection,
		"",
		pathsSection,
		"",
		configSection,
	)

	rows = append(rows, body)

	return styles.BoxStyle.Width(width - 2).Render(
		lipgloss.JoinVertical(lipgloss.Left, rows...),
	)
}
