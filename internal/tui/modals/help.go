package modals

import (
	"skiptui/internal/tui/styles"

	"github.com/charmbracelet/lipgloss"
)

// RenderHelpModal renders the shortcut cheatsheet.
func RenderHelpModal(width int) string {
	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(1, 2).
		Width(70)

	title := styles.TitleStyle.Render("SkipTUI: Keyboard Shortcuts Cheat-sheet")

	row := func(key, desc string) string {
		k := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Width(16).Render(key)
		d := lipgloss.NewStyle().Foreground(styles.ColorFg).Render(desc)
		return lipgloss.JoinHorizontal(lipgloss.Left, k, d)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSecondary).Render("Global Navigation:"),
		row("1 - 4", "Switch Tabs (Sessions, Profiles, Logs, Settings)"),
		row("Tab / Shift+Tab", "Cycle through tabs sequentially"),
		row("j / k / ↑ / ↓", "Navigate up and down lists"),
		row("q / Ctrl+C", "Quit application (prompts Detach vs Kill)"),
		row("?", "Toggle this help cheatsheet"),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSecondary).Render("Sessions View:"),
		row("l", "Open Quick Launcher (spawn app/process)"),
		row("t", "Quick spawn interactive shell in external terminal"),
		row("k / x", "Kill / terminate selected session"),
		row("c", "Clear stopped/dead sessions"),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSecondary).Render("Profiles View:"),
		row("Enter", "Launch terminal shell with selected profile"),
		row("l", "Launch custom app with selected profile"),
		row("t", "Run latency ping test on selected profile"),
		row("T", "Run concurrent latency test on ALL profiles"),
		row("i", "Import .ovpn or WireGuard .conf profile"),
		row("a", "Add new proxy profile (interactive form)"),
		row("d", "Delete selected profile"),
		"",
		lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("[Esc / ?] Close Help"),
	)

	return lipgloss.Place(
		width,
		24,
		lipgloss.Center,
		lipgloss.Center,
		helpBox.Render(content),
	)
}
