package main

import (
	"fmt"
	"skiptui/internal/config"

	"github.com/spf13/cobra"
)

var (
	importCustomName string
	importUsername   string
	importPassword   string
)

var importCmd = &cobra.Command{
	Use:   "import <path/to/profile.ovpn|conf>",
	Short: "Import an OpenVPN (.ovpn) or WireGuard (.conf) configuration file",
	Example: `  skiptui import ~/Downloads/vpnbook-uk205-tcp443.ovpn --name "VPNBook-UK" -u vpnbook -p "secret123"
  skiptui import ~/Downloads/wg0.conf --name "Mullvad-WG"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		filePath := args[0]
		prof, err := cfg.ImportFile(filePath, importCustomName, importUsername, importPassword)
		if err != nil {
			return fmt.Errorf("import failed: %w", err)
		}

		fmt.Printf("✓ Successfully imported %s profile '%s' (ID: %s)\n", prof.Protocol, prof.Name, prof.ID)
		fmt.Printf("  Endpoint: %s | DNS: %s\n", prof.Endpoint, prof.DNS)
		if prof.Username != "" {
			fmt.Printf("  Credentials: User '%s' configured\n", prof.Username)
		}
		return nil
	},
}

func init() {
	importCmd.Flags().StringVarP(&importCustomName, "name", "n", "", "custom name for the imported profile")
	importCmd.Flags().StringVarP(&importUsername, "username", "u", "", "username for OpenVPN authentication")
	importCmd.Flags().StringVarP(&importPassword, "password", "p", "", "password for OpenVPN authentication")
	rootCmd.AddCommand(importCmd)
}
