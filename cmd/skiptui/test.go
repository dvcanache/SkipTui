package main

import (
	"context"
	"fmt"
	"skiptui/internal/config"
	"skiptui/internal/netprobe"
	"time"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test [profile-name-or-id]",
	Short: "Test latency and connectivity to configured proxy and VPN endpoints",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if len(args) > 0 {
			target := args[0]
			p := cfg.GetProfile(target)
			if p == nil {
				return fmt.Errorf("profile '%s' not found", target)
			}
			fmt.Printf("Testing latency to %s (%s)...\n", p.Name, p.Endpoint)
			res := netprobe.TestProfileLatency(context.Background(), p, 3*time.Second)
			if res.Error != nil {
				fmt.Printf("✗ Failed: %v\n", res.Error)
			} else {
				fmt.Printf("✓ Success: %d ms latency\n", res.LatencyMs)
			}
			_ = config.SaveConfig(cfg)
			return nil
		}

		fmt.Printf("Testing %d configured profiles...\n", len(cfg.Profiles))
		results := netprobe.TestAllProfiles(context.Background(), cfg.Profiles, 3*time.Second)
		for _, r := range results {
			p := cfg.GetProfile(r.ProfileID)
			if p != nil {
				if r.Error != nil {
					fmt.Printf("  ✗ %-20s (%-24s): Failed (%v)\n", p.Name, p.Endpoint, r.Error)
				} else {
					fmt.Printf("  ✓ %-20s (%-24s): %d ms\n", p.Name, p.Endpoint, r.LatencyMs)
				}
			}
		}

		_ = config.SaveConfig(cfg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
