package modals

import (
	"skiptui/internal/tui/styles"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// ImportWizardModal represents the file importer state.
type ImportWizardModal struct {
	PathInput textinput.Model
	NameInput textinput.Model
	FocusIdx  int // 0: path, 1: name
	Active    bool
}

func NewImportWizardModal() ImportWizardModal {
	pi := textinput.New()
	pi.Placeholder = "/path/to/profile.ovpn or /path/to/wg0.conf"
	pi.Focus()
	pi.CharLimit = 512
	pi.Width = 48

	ni := textinput.New()
	ni.Placeholder = "Optional custom name (e.g. Work-VPN)"
	ni.CharLimit = 64
	ni.Width = 48

	return ImportWizardModal{
		PathInput: pi,
		NameInput: ni,
		FocusIdx:  0,
		Active:    false,
	}
}

func (m *ImportWizardModal) Render(width int) string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(1, 2).
		Width(64)

	title := styles.TitleStyle.Render("Import VPN Profile (.ovpn or .conf)")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSecondary).Render("File Path:"),
		m.PathInput.View(),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSecondary).Render("Custom Profile Name:"),
		m.NameInput.View(),
		"",
		lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("[Tab] Switch Field    [Enter] Import Profile    [Esc] Cancel"),
	)

	return lipgloss.Place(
		width,
		18,
		lipgloss.Center,
		lipgloss.Center,
		box.Render(content),
	)
}
