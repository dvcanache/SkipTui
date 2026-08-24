package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"skiptui/internal/config"
	"skiptui/internal/isolation"
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

		var engine isolation.Engine = isolation.NewNetnsEngine()
		cfg, _ := config.LoadConfig()
		if cfg != nil && cfg.Settings.RootlessMode {
			engine = isolation.NewRootlessEngine()
		}

		sb := &isolation.SandboxInfo{
			ID:        execSessionID,
			Namespace: "skiptui-" + execSessionID,
		}

		childCmd, err := engine.BuildCommand(context.Background(), sb, targetCmd, targetArgs...)
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
