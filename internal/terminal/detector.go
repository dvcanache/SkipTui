package terminal

import (
	"os"
	"os/exec"
)

// TerminalType represents the detected terminal emulator.
type TerminalType string

const (
	TermKitty         TerminalType = "kitty"
	TermAlacritty     TerminalType = "alacritty"
	TermWezterm       TerminalType = "wezterm"
	TermGhostty       TerminalType = "ghostty"
	TermFoot          TerminalType = "foot"
	TermGnomeTerminal TerminalType = "gnome-terminal"
	TermXterm         TerminalType = "xterm"
	TermTmux          TerminalType = "tmux"
	TermGeneric       TerminalType = "generic"
)

// DetectPreferredTerminal finds the best available terminal emulator on the Linux desktop.
func DetectPreferredTerminal(preference string) (TerminalType, string) {
	// 1. If explicit preference is provided and installed, use it
	if preference != "" && preference != "auto" {
		if path, err := exec.LookPath(preference); err == nil {
			return TerminalType(preference), path
		}
	}

	// 2. Check if running inside tmux
	if os.Getenv("TMUX") != "" {
		if path, err := exec.LookPath("tmux"); err == nil {
			return TermTmux, path
		}
	}

	// 3. Check $TERMINAL environment variable
	if termEnv := os.Getenv("TERMINAL"); termEnv != "" {
		if path, err := exec.LookPath(termEnv); err == nil {
			return TermGeneric, path
		}
	}

	// 4. Probe common modern terminal emulators
	candidates := []TerminalType{
		TermKitty,
		TermAlacritty,
		TermWezterm,
		TermGhostty,
		TermFoot,
		TermGnomeTerminal,
		TermXterm,
	}

	for _, cand := range candidates {
		if path, err := exec.LookPath(string(cand)); err == nil {
			return cand, path
		}
	}

	return TermGeneric, "x-terminal-emulator"
}
