package terminal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// Spawner handles launching isolated commands in a dedicated external terminal window or pane.
type Spawner struct {
	Preference string
}

func NewSpawner(preference string) *Spawner {
	return &Spawner{Preference: preference}
}

// SpawnInExternalTerminal launches a command inside an external terminal emulator window.
func (s *Spawner) SpawnInExternalTerminal(ctx context.Context, sessionID string, binary string, targetCmd string, args []string) error {
	termType, termPath := DetectPreferredTerminal(s.Preference)

	// Format full command to execute
	// e.g. "skiptui exec --session <id> -- <cmd>"
	selfExe, err := os.Executable()
	if err != nil {
		selfExe = "skiptui"
	}

	execArgs := []string{"exec", "--session", sessionID, "--", targetCmd}
	execArgs = append(execArgs, args...)

	var cmd *exec.Cmd

	switch termType {
	case TermTmux:
		tmuxCmd := fmt.Sprintf("%s %s", selfExe, strings.Join(execArgs, " "))
		cmd = exec.CommandContext(ctx, termPath, "split-window", "-h", tmuxCmd)
	case TermKitty:
		var fullArgs []string
		fullArgs = append(fullArgs, "-e", selfExe)
		fullArgs = append(fullArgs, execArgs...)
		cmd = exec.CommandContext(ctx, termPath, fullArgs...)
	case TermAlacritty:
		var fullArgs []string
		fullArgs = append(fullArgs, "-e", selfExe)
		fullArgs = append(fullArgs, execArgs...)
		cmd = exec.CommandContext(ctx, termPath, fullArgs...)
	case TermWezterm:
		var fullArgs []string
		fullArgs = append(fullArgs, "start", "--", selfExe)
		fullArgs = append(fullArgs, execArgs...)
		cmd = exec.CommandContext(ctx, termPath, fullArgs...)
	case TermGhostty:
		var fullArgs []string
		fullArgs = append(fullArgs, "-e", selfExe)
		fullArgs = append(fullArgs, execArgs...)
		cmd = exec.CommandContext(ctx, termPath, fullArgs...)
	case TermFoot:
		var fullArgs []string
		fullArgs = append(fullArgs, selfExe)
		fullArgs = append(fullArgs, execArgs...)
		cmd = exec.CommandContext(ctx, termPath, fullArgs...)
	case TermGnomeTerminal:
		var fullArgs []string
		fullArgs = append(fullArgs, "--", selfExe)
		fullArgs = append(fullArgs, execArgs...)
		cmd = exec.CommandContext(ctx, termPath, fullArgs...)
	default:
		var fullArgs []string
		fullArgs = append(fullArgs, "-e", selfExe)
		fullArgs = append(fullArgs, execArgs...)
		cmd = exec.CommandContext(ctx, termPath, fullArgs...)
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to spawn terminal (%s): %w", termType, err)
	}

	return nil
}
