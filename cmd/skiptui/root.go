package main

import (
	"fmt"
	"os"
	"skiptui/internal/config"
	"skiptui/internal/isolation"
	"skiptui/internal/session"
	"skiptui/internal/tui"
	"skiptui/pkg/version"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "skiptui",
	Short: "SkipTUI - Process-level network isolation and proxy/VPN tunneling TUI for Linux",
	Long: `SkipTUI isolates any application, command, or interactive terminal shell
through a dedicated proxy (SOCKS5, HTTP, Shadowsocks) or VPN (WireGuard, OpenVPN)
while leaving the rest of the host system on its regular network.`,
	Version: fmt.Sprintf("%s (commit: %s, date: %s)", version.Version, version.GitCommit, version.BuildDate),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Sweep any leftover namespaces from previous unclean shutdowns
		_ = isolation.SweepOrphanNamespaces()

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		sup := session.NewSupervisor(cfg)

		// Register graceful shutdown
		isolation.RegisterSignalCleanup(func() {
			// Supervisor handles active cleanup
		})

		return tui.Start(cfg, sup)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/skiptui/config.yaml)")
}
