package cmd

import (
	"context"
	"time"

	"github.com/gradientlinux/concave/internal/config"
	"github.com/gradientlinux/concave/internal/docker"
	"github.com/gradientlinux/concave/internal/suite"
	"github.com/gradientlinux/concave/internal/ui"
	"github.com/gradientlinux/concave/internal/workspace"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [suite]",
	Short: "Start all or one suite",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		names, err := targetSuites(args)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		for _, name := range names {
			if err := docker.ComposeUp(ctx, workspace.ComposePath(name), true); err != nil {
				return err
			}
			ui.Pass("Started", name)
		}
		return nil
	},
}

func targetSuites(args []string) ([]string, error) {
	if len(args) == 1 {
		if _, err := suite.Get(args[0]); err != nil {
			return nil, err
		}
		return []string{args[0]}, nil
	}
	state, err := config.LoadState()
	if err != nil {
		return nil, err
	}
	return state.Installed, nil
}

func init() {
	rootCmd.AddCommand(startCmd)
}
