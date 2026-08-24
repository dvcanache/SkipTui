package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Color Palette (Nord / Modern Terminal Aesthetic)
	ColorPrimary   = lipgloss.Color("#88C0D0") // Frost Blue
	ColorSecondary = lipgloss.Color("#81A1C1") // Sky Blue
	ColorAccent    = lipgloss.Color("#B48EAD") // Aurora Purple
	ColorSuccess   = lipgloss.Color("#A3BE8C") // Aurora Green
	ColorWarning   = lipgloss.Color("#EBCB8B") // Aurora Yellow
	ColorDanger    = lipgloss.Color("#BF616A") // Aurora Red
	ColorBg        = lipgloss.Color("#2E3440") // Polar Night Dark
	ColorFg        = lipgloss.Color("#ECEFF4") // Snow Storm Light
	ColorMuted     = lipgloss.Color("#4C566A") // Polar Night Grey
	ColorBorder    = lipgloss.Color("#3B4252") // Polar Night Mid

	// Common Text Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)

	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorFg).
			Background(ColorSecondary).
			Padding(0, 2)

	TabInactiveStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Background(ColorBg).
				Padding(0, 2)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	BadgeSuccess = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSuccess)

	BadgeDanger = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorDanger)

	BadgeWarning = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWarning)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1)

	SelectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorFg).
				Background(ColorBorder)

	NormalRowStyle = lipgloss.NewStyle().
			Foreground(ColorFg)

	HelpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)
)
