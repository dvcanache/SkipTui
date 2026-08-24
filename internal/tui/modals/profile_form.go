package modals

import (
	"fmt"
	"skiptui/internal/config"
	"skiptui/internal/tui/styles"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// ProfileFormModal manages the interactive creation and editing of proxy/VPN profiles.
type ProfileFormModal struct {
	NameInput       textinput.Model
	HostInput       textinput.Model
	PortInput       textinput.Model
	UserInput       textinput.Model
	PassInput       textinput.Model
	DNSInput        textinput.Model
	Protocols       []config.ProtocolType
	SelectedProto   int
	KillSwitch      bool
	FocusIdx        int // 0: Name, 1: Proto, 2: Host, 3: Port, 4: User, 5: Pass, 6: DNS, 7: KillSwitch, 8: Save
	Active          bool
	IsEditing       bool
	EditingID       string
	ExistingOpenVPN *config.OpenVPNConfig
	ExistingWG      *config.WGConfig
}

func NewProfileFormModal() ProfileFormModal {
	nameIn := textinput.New()
	nameIn.Placeholder = "e.g. US-Residential"
	nameIn.Focus()
	nameIn.CharLimit = 32
	nameIn.Width = 36

	hostIn := textinput.New()
	hostIn.Placeholder = "e.g. 198.51.100.22 or 127.0.0.1"
	hostIn.CharLimit = 64
	hostIn.Width = 36

	portIn := textinput.New()
	portIn.Placeholder = "1080 (or 8080, 51820)"
	portIn.CharLimit = 6
	portIn.Width = 36

	userIn := textinput.New()
	userIn.Placeholder = "Optional username"
	userIn.CharLimit = 32
	userIn.Width = 36

	passIn := textinput.New()
	passIn.Placeholder = "Optional password"
	passIn.EchoMode = textinput.EchoPassword
	passIn.EchoCharacter = '•'
	passIn.CharLimit = 64
	passIn.Width = 36

	dnsIn := textinput.New()
	dnsIn.Placeholder = "1.1.1.1"
	dnsIn.SetValue("1.1.1.1")
	dnsIn.CharLimit = 32
	dnsIn.Width = 36

	return ProfileFormModal{
		NameInput:     nameIn,
		HostInput:     hostIn,
		PortInput:     portIn,
		UserInput:     userIn,
		PassInput:     passIn,
		DNSInput:      dnsIn,
		Protocols:     []config.ProtocolType{config.ProtocolSOCKS5, config.ProtocolHTTP, config.ProtocolShadowsocks, config.ProtocolWireGuard, config.ProtocolOpenVPN},
		SelectedProto: 0,
		KillSwitch:    true,
		FocusIdx:      0,
		Active:        false,
	}
}

// LoadProfile populates the form with an existing profile's values for editing.
func (m *ProfileFormModal) LoadProfile(p *config.Profile) {
	m.IsEditing = true
	m.EditingID = p.ID
	m.ExistingOpenVPN = p.OpenVPN
	m.ExistingWG = p.WireGuard
	m.NameInput.SetValue(p.Name)

	endpointParts := strings.Split(p.Endpoint, ":")
	if len(endpointParts) == 2 {
		m.HostInput.SetValue(endpointParts[0])
		m.PortInput.SetValue(endpointParts[1])
	} else {
		m.HostInput.SetValue(p.Endpoint)
		m.PortInput.SetValue("1080")
	}

	m.UserInput.SetValue(p.Username)
	m.PassInput.SetValue(p.Password)
	if p.DNS != "" {
		m.DNSInput.SetValue(p.DNS)
	} else {
		m.DNSInput.SetValue("1.1.1.1")
	}
	m.KillSwitch = p.KillSwitch

	for i, proto := range m.Protocols {
		if proto == p.Protocol {
			m.SelectedProto = i
			break
		}
	}
	m.FocusIdx = 0
	m.updateFocus()
}

// Reset clears the form inputs.
func (m *ProfileFormModal) Reset() {
	m.IsEditing = false
	m.EditingID = ""
	m.ExistingOpenVPN = nil
	m.ExistingWG = nil
	m.NameInput.SetValue("")
	m.HostInput.SetValue("")
	m.PortInput.SetValue("1080")
	m.UserInput.SetValue("")
	m.PassInput.SetValue("")
	m.DNSInput.SetValue("1.1.1.1")
	m.SelectedProto = 0
	m.KillSwitch = true
	m.FocusIdx = 0
	m.updateFocus()
}

func (m *ProfileFormModal) updateFocus() {
	m.NameInput.Blur()
	m.HostInput.Blur()
	m.PortInput.Blur()
	m.UserInput.Blur()
	m.PassInput.Blur()
	m.DNSInput.Blur()

	switch m.FocusIdx {
	case 0:
		m.NameInput.Focus()
	case 2:
		m.HostInput.Focus()
	case 3:
		m.PortInput.Focus()
	case 4:
		m.UserInput.Focus()
	case 5:
		m.PassInput.Focus()
	case 6:
		m.DNSInput.Focus()
	}
}

// NextField moves focus down to the next input field.
func (m *ProfileFormModal) NextField() {
	m.FocusIdx = (m.FocusIdx + 1) % 9
	m.updateFocus()
}

// PrevField moves focus up to the previous input field.
func (m *ProfileFormModal) PrevField() {
	if m.FocusIdx == 0 {
		m.FocusIdx = 8
	} else {
		m.FocusIdx--
	}
	m.updateFocus()
}

// CycleProtocol switches the selected protocol.
func (m *ProfileFormModal) CycleProtocol(forward bool) {
	if forward {
		m.SelectedProto = (m.SelectedProto + 1) % len(m.Protocols)
	} else {
		if m.SelectedProto == 0 {
			m.SelectedProto = len(m.Protocols) - 1
		} else {
			m.SelectedProto--
		}
	}

	if m.PortInput.Value() == "" || m.PortInput.Value() == "1080" || m.PortInput.Value() == "8080" || m.PortInput.Value() == "1194" || m.PortInput.Value() == "51820" {
		switch m.Protocols[m.SelectedProto] {
		case config.ProtocolSOCKS5:
			m.PortInput.SetValue("1080")
		case config.ProtocolHTTP:
			m.PortInput.SetValue("8080")
		case config.ProtocolWireGuard:
			m.PortInput.SetValue("51820")
		case config.ProtocolOpenVPN:
			m.PortInput.SetValue("1194")
		}
	}
}

// ToProfile builds a config.Profile from the form inputs, preserving existing configurations.
func (m *ProfileFormModal) ToProfile() *config.Profile {
	name := strings.TrimSpace(m.NameInput.Value())
	if name == "" {
		name = fmt.Sprintf("Proxy-%s", string(m.Protocols[m.SelectedProto]))
	}

	host := strings.TrimSpace(m.HostInput.Value())
	if host == "" {
		host = "127.0.0.1"
	}

	port := strings.TrimSpace(m.PortInput.Value())
	if port == "" {
		port = "1080"
	}

	endpoint := fmt.Sprintf("%s:%s", host, port)

	dns := strings.TrimSpace(m.DNSInput.Value())
	if dns == "" {
		dns = "1.1.1.1"
	}

	id := m.EditingID
	if id == "" {
		id = fmt.Sprintf("p-%d", time.Now().Unix()%100000)
	}

	return &config.Profile{
		ID:         id,
		Name:       name,
		Protocol:   m.Protocols[m.SelectedProto],
		Endpoint:   endpoint,
		Username:   strings.TrimSpace(m.UserInput.Value()),
		Password:   strings.TrimSpace(m.PassInput.Value()),
		DNS:        dns,
		KillSwitch: m.KillSwitch,
		WireGuard:  m.ExistingWG,
		OpenVPN:    m.ExistingOpenVPN,
		CreatedAt:  time.Now(),
	}
}

// Render draws the interactive profile configuration form.
func (m *ProfileFormModal) Render(width int) string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(1, 2).
		Width(68)

	titleText := "Configure Network Proxy / VPN Endpoint"
	if m.IsEditing {
		titleText = "Edit Proxy / VPN Profile: " + m.NameInput.Value()
	}
	title := styles.TitleStyle.Render(titleText)

	var protoPills []string
	for i, proto := range m.Protocols {
		pStr := string(proto)
		if i == m.SelectedProto {
			protoPills = append(protoPills, lipgloss.NewStyle().Bold(true).Foreground(styles.ColorFg).Background(styles.ColorSecondary).Padding(0, 1).Render(pStr))
		} else {
			protoPills = append(protoPills, lipgloss.NewStyle().Foreground(styles.ColorMuted).Padding(0, 1).Render(pStr))
		}
	}
	protoSelector := lipgloss.JoinHorizontal(lipgloss.Left, protoPills...)
	if m.FocusIdx == 1 {
		protoSelector = lipgloss.NewStyle().Bold(true).Border(lipgloss.NormalBorder()).BorderForeground(styles.ColorPrimary).Render(protoSelector + "  (← / → to change)")
	}

	fieldLabel := func(label string, idx int) string {
		style := lipgloss.NewStyle().Width(18).Foreground(styles.ColorFg)
		if m.FocusIdx == idx {
			style = style.Bold(true).Foreground(styles.ColorPrimary)
		}
		return style.Render(label)
	}

	row := func(label string, idx int, inputView string) string {
		return lipgloss.JoinHorizontal(lipgloss.Left, fieldLabel(label, idx), inputView)
	}

	ksCheckbox := "[ ] Disabled"
	if m.KillSwitch {
		ksCheckbox = "[X] Enabled (Fail-Closed KillSwitch)"
	}
	if m.FocusIdx == 7 {
		ksCheckbox = lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary).Render(ksCheckbox + " (Space/Tab to toggle)")
	}

	saveBtn := "[ Save & Add Profile ]"
	if m.IsEditing {
		saveBtn = "[ Save Changes ]"
	}
	if m.FocusIdx == 8 {
		saveBtn = lipgloss.NewStyle().Bold(true).Foreground(styles.ColorFg).Background(styles.ColorSuccess).Padding(0, 2).Render(saveBtn)
	} else {
		saveBtn = lipgloss.NewStyle().Foreground(styles.ColorSuccess).Padding(0, 2).Render(saveBtn)
	}

	cancelBtn := lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("[Esc] Cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		row("Profile Name:", 0, m.NameInput.View()),
		row("Protocol Type:", 1, protoSelector),
		row("IP Address / Host:", 2, m.HostInput.View()),
		row("Port Number:", 3, m.PortInput.View()),
		row("Username (opt):", 4, m.UserInput.View()),
		row("Password (opt):", 5, m.PassInput.View()),
		row("DNS Server:", 6, m.DNSInput.View()),
		row("Leak KillSwitch:", 7, ksCheckbox),
		"",
		lipgloss.JoinHorizontal(lipgloss.Left, saveBtn, "   ", cancelBtn),
		"",
		lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("[Tab / ↓] Next Field   [Shift+Tab / ↑] Prev Field   [Enter] Save"),
	)

	return lipgloss.Place(
		width,
		26,
		lipgloss.Center,
		lipgloss.Center,
		box.Render(content),
	)
}
