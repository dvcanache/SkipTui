package modals

import (
	"skiptui/internal/tui/styles"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// ImportWizardModal represents the file importer state with credentials.
type ImportWizardModal struct {
	PathInput textinput.Model
	NameInput textinput.Model
	UserInput textinput.Model
	PassInput textinput.Model
	FocusIdx  int // 0: path, 1: name, 2: username, 3: password, 4: import
	Active    bool
}

func NewImportWizardModal() ImportWizardModal {
	pi := textinput.New()
	pi.Placeholder = "/path/to/vpnbook-uk205-tcp443.ovpn"
	pi.Focus()
	pi.CharLimit = 512
	pi.Width = 44

	ni := textinput.New()
	ni.Placeholder = "Optional name (e.g. VPNBook-UK)"
	ni.CharLimit = 64
	ni.Width = 44

	ui := textinput.New()
	ui.Placeholder = "Optional username (e.g. vpnbook)"
	ui.CharLimit = 64
	ui.Width = 44

	pwd := textinput.New()
	pwd.Placeholder = "Optional password"
	pwd.EchoMode = textinput.EchoPassword
	pwd.EchoCharacter = '•'
	pwd.CharLimit = 64
	pwd.Width = 44

	return ImportWizardModal{
		PathInput: pi,
		NameInput: ni,
		UserInput: ui,
		PassInput: pwd,
		FocusIdx:  0,
		Active:    false,
	}
}

func (m *ImportWizardModal) NextField() {
	m.FocusIdx = (m.FocusIdx + 1) % 5
	m.updateFocus()
}

func (m *ImportWizardModal) PrevField() {
	if m.FocusIdx == 0 {
		m.FocusIdx = 4
	} else {
		m.FocusIdx--
	}
	m.updateFocus()
}

func (m *ImportWizardModal) updateFocus() {
	m.PathInput.Blur()
	m.NameInput.Blur()
	m.UserInput.Blur()
	m.PassInput.Blur()

	switch m.FocusIdx {
	case 0:
		m.PathInput.Focus()
	case 1:
		m.NameInput.Focus()
	case 2:
		m.UserInput.Focus()
	case 3:
		m.PassInput.Focus()
	}
}

func (m *ImportWizardModal) Render(width int) string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(1, 2).
		Width(66)

	title := styles.TitleStyle.Render("Import OpenVPN (.ovpn) or WireGuard (.conf)")

	fieldLabel := func(label string, idx int) string {
		style := lipgloss.NewStyle().Width(16).Foreground(styles.ColorFg)
		if m.FocusIdx == idx {
			style = style.Bold(true).Foreground(styles.ColorPrimary)
		}
		return style.Render(label)
	}

	row := func(label string, idx int, inputView string) string {
		return lipgloss.JoinHorizontal(lipgloss.Left, fieldLabel(label, idx), inputView)
	}

	importBtn := "[ Import & Save Profile ]"
	if m.FocusIdx == 4 {
		importBtn = lipgloss.NewStyle().Bold(true).Foreground(styles.ColorFg).Background(styles.ColorSuccess).Padding(0, 2).Render(importBtn)
	} else {
		importBtn = lipgloss.NewStyle().Foreground(styles.ColorSuccess).Padding(0, 2).Render(importBtn)
	}

	cancelBtn := lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("[Esc] Cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		row("File Path:", 0, m.PathInput.View()),
		row("Profile Name:", 1, m.NameInput.View()),
		row("Username (opt):", 2, m.UserInput.View()),
		row("Password (opt):", 3, m.PassInput.View()),
		"",
		lipgloss.JoinHorizontal(lipgloss.Left, importBtn, "   ", cancelBtn),
		"",
		lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("[Tab / ↓] Next Field   [Shift+Tab / ↑] Prev Field   [Enter] Import"),
	)

	return lipgloss.Place(
		width,
		20,
		lipgloss.Center,
		lipgloss.Center,
		box.Render(content),
	)
}
