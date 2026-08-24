package main

import (
	"context"
	"fmt"
	"skiptui/internal/config"
	"skiptui/internal/session"

	"github.com/spf13/cobra"
)

var killCmd = &cobra.Command{
	Use:   "kill <session-id>",
	Short: "Terminate an active isolated sandbox session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		sessionID := args[0]
		sup := session.NewSupervisor(cfg)

		if err := sup.KillSession(context.Background(), sessionID); err != nil {
			return fmt.Errorf("failed to kill session %s: %w", sessionID, err)
		}

		fmt.Printf("✓ Successfully killed session %s\n", sessionID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(killCmd)
}
