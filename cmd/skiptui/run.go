package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"skiptui/internal/config"
	"skiptui/internal/isolation"
	"skiptui/internal/session"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	runProfileName string
	runInTerminal  bool
)

var runCmd = &cobra.Command{
	Use:   "run [flags] -- <command> [args...]",
	Short: "Launch a process or terminal shell inside an isolated proxy/VPN sandbox",
	Example: `  skiptui run --profile us-residential -- firefox
  skiptui run --profile nl-wireguard --terminal -- zsh
  skiptui run --profile local-tor -- curl https://api.ipify.org`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Find requested profile or fallback to default
		var profile *config.Profile
		if runProfileName != "" {
			profile = cfg.GetProfile(runProfileName)
			if profile == nil {
				return fmt.Errorf("profile '%s' not found", runProfileName)
			}
		} else if cfg.Settings.DefaultProfile != "" {
			profile = cfg.GetProfile(cfg.Settings.DefaultProfile)
		}

		if profile == nil {
			if len(cfg.Profiles) > 0 {
				profile = cfg.Profiles[0]
			} else {
				return fmt.Errorf("no profiles configured; use 'skiptui import' or 'skiptui' to add one")
			}
		}

		targetCmd := args[0]
		targetArgs := args[1:]

		sup := session.NewSupervisor(cfg)

		if runInTerminal {
			sess, err := sup.LaunchSession(context.Background(), profile, targetCmd, targetArgs, true)
			if err != nil {
				return fmt.Errorf("failed to launch session: %w", err)
			}
			fmt.Printf("✓ Launched session %s (%s) in external terminal\n", sess.ID, profile.Name)
			return nil
		}

		// Pick isolation engine
		engine := sup.SelectBestEngine()

		sb, err := engine.CreateSandbox(context.Background(), "cli-exec", profile)
		if err != nil {
			// Fallback to EnvProxyEngine
			engine = isolation.NewEnvProxyEngine()
			sb, err = engine.CreateSandbox(context.Background(), "cli-exec", profile)
			if err != nil {
				return fmt.Errorf("failed to create sandbox: %w", err)
			}
		}
		defer func() {
			_ = engine.DestroySandbox(context.Background(), sb)
		}()

		var execCmd *exec.Cmd
		if envEng, ok := engine.(*isolation.EnvProxyEngine); ok {
			execCmd, err = envEng.BuildCommandWithProfile(context.Background(), sb, profile, targetCmd, targetArgs...)
		} else {
			execCmd, err = engine.BuildCommand(context.Background(), sb, targetCmd, targetArgs...)
		}

		if err != nil {
			return fmt.Errorf("failed to build command: %w", err)
		}

		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr

		if err := execCmd.Run(); err != nil {
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
	runCmd.Flags().StringVarP(&runProfileName, "profile", "p", "", "name or ID of the profile to use")
	runCmd.Flags().BoolVarP(&runInTerminal, "terminal", "t", false, "launch command in a new external terminal window")
	rootCmd.AddCommand(runCmd)
}
