package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"skiptui/internal/config"
	"skiptui/internal/isolation"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	execSessionID string
)

var execCmd = &cobra.Command{
	Use:    "exec [flags] -- <command> [args...]",
	Short:  "Internal helper to execute a command inside a specific sandbox namespace",
	Hidden: true,
	Args:   cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetCmd := args[0]
		targetArgs := args[1:]

		cfg, _ := config.LoadConfig()

		var engine isolation.Engine = isolation.NewNetnsEngine()
		if strings.HasPrefix(execSessionID, "env-") || (cfg != nil && cfg.Settings.RootlessMode) || engine.CheckCapabilities() != nil {
			engine = isolation.NewEnvProxyEngine()
		}

		sb := &isolation.SandboxInfo{
			ID:        execSessionID,
			Namespace: "skiptui-" + execSessionID,
		}

		var childCmd *exec.Cmd
		var err error

		if envEng, ok := engine.(*isolation.EnvProxyEngine); ok {
			var prof *config.Profile
			if cfg != nil && len(cfg.Profiles) > 0 {
				prof = cfg.Profiles[0]
			}
			childCmd, err = envEng.BuildCommandWithProfile(context.Background(), sb, prof, targetCmd, targetArgs...)
		} else {
			childCmd, err = engine.BuildCommand(context.Background(), sb, targetCmd, targetArgs...)
		}

		if err != nil {
			return fmt.Errorf("failed to build command: %w", err)
		}

		childCmd.Stdin = os.Stdin
		childCmd.Stdout = os.Stdout
		childCmd.Stderr = os.Stderr

		if err := childCmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					os.Exit(status.ExitStatus())
				}
			}
			return err
		}

		return nil
	},
}

func init() {
	execCmd.Flags().StringVar(&execSessionID, "session", "", "session ID")
	rootCmd.AddCommand(execCmd)
}
