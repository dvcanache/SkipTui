package main

import (
	"fmt"
	"skiptui/internal/config"

	"github.com/spf13/cobra"
)

var (
	importCustomName string
)

var importCmd = &cobra.Command{
	Use:   "import <path/to/profile.ovpn|conf>",
	Short: "Import an OpenVPN (.ovpn) or WireGuard (.conf) configuration file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		filePath := args[0]
		prof, err := cfg.ImportFile(filePath, importCustomName)
		if err != nil {
			return fmt.Errorf("import failed: %w", err)
		}

		fmt.Printf("✓ Successfully imported %s profile '%s' (ID: %s)\n", prof.Protocol, prof.Name, prof.ID)
		fmt.Printf("  Endpoint: %s | DNS: %s\n", prof.Endpoint, prof.DNS)
		return nil
	},
}

func init() {
	importCmd.Flags().StringVarP(&importCustomName, "name", "n", "", "custom name for the imported profile")
	rootCmd.AddCommand(importCmd)
}
