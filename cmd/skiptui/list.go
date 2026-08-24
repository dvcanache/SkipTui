package main

import (
	"fmt"
	"os"
	"skiptui/internal/config"
	"skiptui/internal/daemon"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List active isolated sessions or configured profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		sessions, _ := daemon.LoadSessionState()
		// Validate active liveness
		for _, s := range sessions {
			if s.Status == "running" && s.PID > 0 && !daemon.IsPIDAlive(s.PID) {
				s.Status = "stopped"
			}
		}
		_ = daemon.SaveSessionState(sessions)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

		fmt.Println("--- CONFIGURED PROFILES ---")
		fmt.Fprintln(w, "NAME\tPROTOCOL\tENDPOINT\tDNS\tKILLSWITCH")
		for _, p := range cfg.Profiles {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%v\n", p.Name, p.Protocol, p.Endpoint, p.DNS, p.KillSwitch)
		}
		w.Flush()

		fmt.Println("\n--- ACTIVE SESSIONS ---")
		if len(sessions) == 0 {
			fmt.Println("No active isolated sessions running.")
		} else {
			fmt.Fprintln(w, "ID\tCOMMAND\tPROFILE\tSTATUS\tPID\tNAMESPACE")
			for _, s := range sessions {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n", s.ID, s.Command, s.ProfileName, s.Status, s.PID, s.Namespace)
			}
			w.Flush()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
