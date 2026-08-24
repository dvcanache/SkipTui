package tui

import (
	"fmt"
	"skiptui/internal/tui/modals"
	"skiptui/internal/tui/styles"
	"skiptui/internal/tui/views"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.Width == 0 {
		m.Width = 80
	}

	// 1. Modals Overlay
	if m.ShowExitDialog {
		return modals.RenderExitDialog(len(m.Sessions), m.Width)
	}

	if m.ShowHelp {
		return modals.RenderHelpModal(m.Width)
	}

	if m.ShowLauncher {
		return m.Launcher.Render(m.Width)
	}

	if m.ShowImporter {
		return m.Importer.Render(m.Width)
	}

	// 2. Top Header & Tabs
	title := styles.TitleStyle.Render("SkipTUI v0.1.0 🦘")

	tabs := []string{"[1] Sessions", "[2] Profiles", "[3] Logs", "[4] Settings"}
	var renderedTabs []string
	for i, tab := range tabs {
		if TabType(i) == m.ActiveTab {
			renderedTabs = append(renderedTabs, styles.TabActiveStyle.Render(tab))
		} else {
			renderedTabs = append(renderedTabs, styles.TabInactiveStyle.Render(tab))
		}
	}
	tabsRow := lipgloss.JoinHorizontal(lipgloss.Left, renderedTabs...)
	topBar := lipgloss.JoinHorizontal(lipgloss.Left, title, "  ", tabsRow)

	// 3. Main Body View
	var body string
	switch m.ActiveTab {
	case TabSessions:
		body = views.RenderDashboard(m.Sessions, m.SelectedSess, m.Width)
	case TabProfiles:
		body = views.RenderProfiles(m.Profiles, m.SelectedProf, m.Width)
	case TabLogs:
		body = views.RenderLogs(m.Logs, m.Width)
	case TabSettings:
		body = views.RenderSettings(m.Config, m.Width)
	}

	// 4. Bottom Status & Hotkey Bar
	status := ""
	if m.StatusMsg != "" {
		status = lipgloss.NewStyle().Foreground(styles.ColorWarning).Render(" " + m.StatusMsg + " |")
	}

	hotkeys := lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(
		fmt.Sprintf("%s [l] Launch  [t] Terminal  [i] Import  [k/x] Kill  [?] Help  [q] Quit", status),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		topBar,
		"",
		body,
		"",
		hotkeys,
	)
}
