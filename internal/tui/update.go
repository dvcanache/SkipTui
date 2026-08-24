package tui

import (
	"context"
	"fmt"
	"skiptui/internal/config"
	"skiptui/internal/netprobe"
	"skiptui/internal/tui/modals"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case TickMsg:
		m.Sessions = m.Supervisor.ListSessions()
		return m, tickEvery(1 * time.Second)

	case LatencyResultMsg:
		for _, p := range m.Profiles {
			if p.ID == msg.ProfileID {
				p.LatencyMs = msg.LatencyMs
				p.LastTested = time.Now()
				break
			}
		}
		_ = config.SaveConfig(m.Config)
		return m, nil

	case tea.KeyMsg:
		// 1. Handle Exit Dialog
		if m.ShowExitDialog {
			switch msg.String() {
			case "d", "D":
				return m, tea.Quit
			case "k", "K":
				m.Supervisor.CleanupAll(context.Background())
				return m, tea.Quit
			case "esc":
				m.ShowExitDialog = false
				return m, nil
			}
			return m, nil
		}

		// 2. Handle Help Modal
		if m.ShowHelp {
			if msg.String() == "esc" || msg.String() == "?" {
				m.ShowHelp = false
			}
			return m, nil
		}

		// 3. Handle Terminal Profile Picker Modal [t]
		if m.ShowTerminalPicker {
			switch msg.String() {
			case "esc":
				m.ShowTerminalPicker = false
				return m, nil
			case "up", "k":
				m.TerminalPicker.Prev()
				return m, nil
			case "down", "j":
				m.TerminalPicker.Next()
				return m, nil
			case "1", "2", "3", "4", "5", "6", "7", "8", "9":
				idx := int(msg.String()[0] - '1')
				if idx >= 0 && idx < len(m.TerminalPicker.Profiles) {
					m.TerminalPicker.SelectedIdx = idx
				}
				return m, nil
			case "enter":
				prof := m.TerminalPicker.GetSelectedProfile()
				if prof != nil {
					sess, err := m.Supervisor.LaunchSession(context.Background(), prof, m.TerminalPicker.Shell, nil, true)
					if err != nil {
						m.StatusMsg = fmt.Sprintf("Failed to spawn terminal: %v", err)
						m.Logs = append(m.Logs, fmt.Sprintf("[%s] Error spawning terminal: %v", time.Now().Format("15:04:05"), err))
					} else {
						m.Sessions = m.Supervisor.ListSessions()
						m.StatusMsg = fmt.Sprintf("✓ Spawned '%s' terminal in session %s (%s)", m.TerminalPicker.Shell, sess.ID, prof.Name)
						m.Logs = append(m.Logs, fmt.Sprintf("[%s] Launched isolated %s terminal with profile %s (%s)", time.Now().Format("15:04:05"), m.TerminalPicker.Shell, prof.Name, prof.Protocol))
						m.ActiveTab = TabSessions
					}
				}
				m.ShowTerminalPicker = false
				return m, nil
			}
			return m, nil
		}

		// 4. Handle Interactive Profile Add/Edit Form
		if m.ShowProfileForm {
			switch msg.String() {
			case "esc":
				m.ShowProfileForm = false
				return m, nil
			case "tab", "down":
				m.ProfileForm.NextField()
				return m, nil
			case "shift+tab", "up":
				m.ProfileForm.PrevField()
				return m, nil
			case "left":
				if m.ProfileForm.FocusIdx == 1 {
					m.ProfileForm.CycleProtocol(false)
					return m, nil
				}
			case "right":
				if m.ProfileForm.FocusIdx == 1 {
					m.ProfileForm.CycleProtocol(true)
					return m, nil
				}
			case " ":
				if m.ProfileForm.FocusIdx == 7 {
					m.ProfileForm.KillSwitch = !m.ProfileForm.KillSwitch
					return m, nil
				}
			case "enter":
				if m.ProfileForm.FocusIdx == 8 || m.ProfileForm.FocusIdx == 0 || m.ProfileForm.FocusIdx == 6 {
					p := m.ProfileForm.ToProfile()
					if err := m.Config.AddProfile(p); err != nil {
						m.StatusMsg = fmt.Sprintf("Error saving profile: %v", err)
					} else {
						m.Profiles = m.Config.Profiles
						m.Launcher.Profiles = m.Profiles
						m.TerminalPicker.Profiles = m.Profiles
						m.StatusMsg = fmt.Sprintf("✓ Saved profile '%s' (%s)", p.Name, p.Endpoint)
						m.Logs = append(m.Logs, fmt.Sprintf("[%s] Configured %s profile: %s (%s)", time.Now().Format("15:04:05"), p.Protocol, p.Name, p.Endpoint))

						targetP := p
						cmds = append(cmds, func() tea.Msg {
							res := netprobe.TestProfileLatency(context.Background(), targetP, 3*time.Second)
							return LatencyResultMsg{
								ProfileID: res.ProfileID,
								LatencyMs: res.LatencyMs,
								Error:     res.Error,
							}
						})
					}
					m.ShowProfileForm = false
					return m, tea.Batch(cmds...)
				} else {
					m.ProfileForm.NextField()
					return m, nil
				}
			default:
				var cmd tea.Cmd
				switch m.ProfileForm.FocusIdx {
				case 0:
					m.ProfileForm.NameInput, cmd = m.ProfileForm.NameInput.Update(msg)
				case 2:
					m.ProfileForm.HostInput, cmd = m.ProfileForm.HostInput.Update(msg)
				case 3:
					m.ProfileForm.PortInput, cmd = m.ProfileForm.PortInput.Update(msg)
				case 4:
					m.ProfileForm.UserInput, cmd = m.ProfileForm.UserInput.Update(msg)
				case 5:
					m.ProfileForm.PassInput, cmd = m.ProfileForm.PassInput.Update(msg)
				case 6:
					m.ProfileForm.DNSInput, cmd = m.ProfileForm.DNSInput.Update(msg)
				}
				return m, cmd
			}
		}

		// 5. Handle Importer Modal (with Credentials)
		if m.ShowImporter {
			switch msg.String() {
			case "esc":
				m.ShowImporter = false
				return m, nil
			case "tab", "down":
				m.Importer.NextField()
				return m, nil
			case "shift+tab", "up":
				m.Importer.PrevField()
				return m, nil
			case "enter":
				if m.Importer.FocusIdx == 4 || m.Importer.FocusIdx == 0 {
					filePath := m.Importer.PathInput.Value()
					customName := m.Importer.NameInput.Value()
					user := m.Importer.UserInput.Value()
					pwd := m.Importer.PassInput.Value()

					if filePath != "" {
						prof, err := m.Config.ImportFile(filePath, customName, user, pwd)
						if err != nil {
							m.StatusMsg = fmt.Sprintf("Import failed: %v", err)
						} else {
							m.Profiles = m.Config.Profiles
							m.Launcher.Profiles = m.Profiles
							m.TerminalPicker.Profiles = m.Profiles
							m.StatusMsg = fmt.Sprintf("✓ Imported %s profile '%s'", prof.Protocol, prof.Name)
							m.Logs = append(m.Logs, fmt.Sprintf("[%s] Imported %s profile: %s (%s)", time.Now().Format("15:04:05"), prof.Protocol, prof.Name, prof.Endpoint))

							targetP := prof
							cmds = append(cmds, func() tea.Msg {
								res := netprobe.TestProfileLatency(context.Background(), targetP, 3*time.Second)
								return LatencyResultMsg{
									ProfileID: res.ProfileID,
									LatencyMs: res.LatencyMs,
									Error:     res.Error,
								}
							})
						}
					}
					m.ShowImporter = false
					return m, tea.Batch(cmds...)
				} else {
					m.Importer.NextField()
					return m, nil
				}
			default:
				var cmd tea.Cmd
				switch m.Importer.FocusIdx {
				case 0:
					m.Importer.PathInput, cmd = m.Importer.PathInput.Update(msg)
				case 1:
					m.Importer.NameInput, cmd = m.Importer.NameInput.Update(msg)
				case 2:
					m.Importer.UserInput, cmd = m.Importer.UserInput.Update(msg)
				case 3:
					m.Importer.PassInput, cmd = m.Importer.PassInput.Update(msg)
				}
				return m, cmd
			}
		}

		// 6. Handle Launcher Modal
		if m.ShowLauncher {
			switch msg.String() {
			case "esc":
				m.ShowLauncher = false
				return m, nil
			case "up":
				if m.Launcher.SelectedProf > 0 {
					m.Launcher.SelectedProf--
				}
				return m, nil
			case "down":
				if m.Launcher.SelectedProf < len(m.Launcher.Profiles)-1 {
					m.Launcher.SelectedProf++
				}
				return m, nil
			case "tab":
				m.Launcher.InTerminal = !m.Launcher.InTerminal
				return m, nil
			case "enter":
				cmdStr := m.Launcher.CommandInput.Value()
				if cmdStr == "" {
					cmdStr = "zsh"
				}
				if len(m.Launcher.Profiles) > 0 {
					prof := m.Launcher.Profiles[m.Launcher.SelectedProf]
					sess, err := m.Supervisor.LaunchSession(context.Background(), prof, cmdStr, nil, m.Launcher.InTerminal)
					if err != nil {
						m.StatusMsg = fmt.Sprintf("Launch failed: %v", err)
						m.Logs = append(m.Logs, fmt.Sprintf("[%s] Error: %v", time.Now().Format("15:04:05"), err))
					} else {
						m.Sessions = m.Supervisor.ListSessions()
						m.StatusMsg = fmt.Sprintf("Session %s launched with profile '%s'", sess.ID, prof.Name)
						m.Logs = append(m.Logs, fmt.Sprintf("[%s] Launched %s in sandbox %s (%s)", time.Now().Format("15:04:05"), cmdStr, sess.ID, prof.Name))
					}
				}
				m.ShowLauncher = false
				return m, nil
			default:
				var cmd tea.Cmd
				m.Launcher.CommandInput, cmd = m.Launcher.CommandInput.Update(msg)
				return m, cmd
			}
		}

		// 7. Global Hotkeys & Screen Navigation
		switch msg.String() {
		case "q", "ctrl+c":
			if len(m.Sessions) > 0 {
				m.ShowExitDialog = true
				return m, nil
			}
			return m, tea.Quit

		case "?":
			m.ShowHelp = true
			return m, nil

		case "1":
			m.ActiveTab = TabSessions
			return m, nil
		case "2":
			m.ActiveTab = TabProfiles
			return m, nil
		case "3":
			m.ActiveTab = TabLogs
			return m, nil
		case "4":
			m.ActiveTab = TabSettings
			return m, nil

		case "tab":
			m.ActiveTab = (m.ActiveTab + 1) % 4
			return m, nil
		case "shift+tab":
			if m.ActiveTab == 0 {
				m.ActiveTab = 3
			} else {
				m.ActiveTab--
			}
			return m, nil

		case "j", "down":
			if m.ActiveTab == TabSessions && len(m.Sessions) > 0 {
				if m.SelectedSess < len(m.Sessions)-1 {
					m.SelectedSess++
				}
			} else if m.ActiveTab == TabProfiles && len(m.Profiles) > 0 {
				if m.SelectedProf < len(m.Profiles)-1 {
					m.SelectedProf++
				}
			}
			return m, nil

		case "k", "up":
			if m.ActiveTab == TabSessions && len(m.Sessions) > 0 {
				if m.SelectedSess > 0 {
					m.SelectedSess--
				}
			} else if m.ActiveTab == TabProfiles && len(m.Profiles) > 0 {
				if m.SelectedProf > 0 {
					m.SelectedProf--
				}
			}
			return m, nil

		case "t":
			// Open the dedicated Terminal Profile Selection Menu!
			m.TerminalPicker = modals.NewTerminalPickerModal(m.Profiles)
			if m.ActiveTab == TabProfiles && m.SelectedProf < len(m.Profiles) {
				m.TerminalPicker.SelectedIdx = m.SelectedProf
			}
			m.ShowTerminalPicker = true
			return m, nil

		case "a":
			m.ProfileForm.Reset()
			m.ShowProfileForm = true
			return m, textinput.Blink

		case "e":
			if m.ActiveTab == TabProfiles && len(m.Profiles) > 0 && m.SelectedProf < len(m.Profiles) {
				p := m.Profiles[m.SelectedProf]
				m.ProfileForm.LoadProfile(p)
				m.ShowProfileForm = true
				return m, textinput.Blink
			}
			return m, nil

		case "l":
			m.Launcher = modals.NewLauncherModal(m.Profiles)
			m.Launcher.CommandInput.Focus()
			m.ShowLauncher = true
			return m, textinput.Blink

		case "i":
			m.Importer = modals.NewImportWizardModal()
			m.Importer.PathInput.Focus()
			m.ShowImporter = true
			return m, textinput.Blink

		case "p", "P":
			// Latency probe hotkey
			if m.ActiveTab == TabProfiles && len(m.Profiles) > 0 && m.SelectedProf < len(m.Profiles) {
				p := m.Profiles[m.SelectedProf]
				m.StatusMsg = fmt.Sprintf("Testing latency for '%s'...", p.Name)
				return m, func() tea.Msg {
					res := netprobe.TestProfileLatency(context.Background(), p, 3*time.Second)
					return LatencyResultMsg{
						ProfileID: res.ProfileID,
						LatencyMs: res.LatencyMs,
						Error:     res.Error,
					}
				}
			}
			return m, nil

		case "T":
			m.StatusMsg = "Testing latency for all profiles..."
			for _, prof := range m.Profiles {
				p := prof
				cmds = append(cmds, func() tea.Msg {
					res := netprobe.TestProfileLatency(context.Background(), p, 3*time.Second)
					return LatencyResultMsg{
						ProfileID: res.ProfileID,
						LatencyMs: res.LatencyMs,
						Error:     res.Error,
					}
				})
			}
			return m, tea.Batch(cmds...)

		case "enter":
			if m.ActiveTab == TabProfiles && len(m.Profiles) > 0 && m.SelectedProf < len(m.Profiles) {
				prof := m.Profiles[m.SelectedProf]
				sess, err := m.Supervisor.LaunchSession(context.Background(), prof, "zsh", nil, true)
				if err != nil {
					m.StatusMsg = fmt.Sprintf("Failed to spawn terminal: %v", err)
				} else {
					m.Sessions = m.Supervisor.ListSessions()
					m.StatusMsg = fmt.Sprintf("Spawned terminal in session %s (%s)", sess.ID, prof.Name)
					m.ActiveTab = TabSessions
				}
			}
			return m, nil

		case "x":
			if m.ActiveTab == TabSessions && len(m.Sessions) > 0 && m.SelectedSess < len(m.Sessions) {
				s := m.Sessions[m.SelectedSess]
				_ = m.Supervisor.KillSession(context.Background(), s.ID)
				m.Sessions = m.Supervisor.ListSessions()
				m.StatusMsg = fmt.Sprintf("Killed session %s", s.ID)
				m.Logs = append(m.Logs, fmt.Sprintf("[%s] Terminated session %s", time.Now().Format("15:04:05"), s.ID))
			}
			return m, nil

		case "d":
			if m.ActiveTab == TabProfiles && len(m.Profiles) > 0 && m.SelectedProf < len(m.Profiles) {
				p := m.Profiles[m.SelectedProf]
				_ = m.Config.DeleteProfile(p.ID)
				m.Profiles = m.Config.Profiles
				m.Launcher.Profiles = m.Profiles
				m.TerminalPicker.Profiles = m.Profiles
				m.StatusMsg = fmt.Sprintf("Deleted profile '%s'", p.Name)
			}
			return m, nil
		}
	}

	return m, nil
}
