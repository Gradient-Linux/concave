package cmd

import (
	"context"
	"time"

	"github.com/gradientlinux/concave/internal/docker"
	"github.com/gradientlinux/concave/internal/ui"
	"github.com/gradientlinux/concave/internal/workspace"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [suite]",
	Short: "Stop all or one suite",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		names, err := targetSuites(args)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		for _, name := range names {
			if err := docker.ComposeDown(ctx, workspace.ComposePath(name)); err != nil {
				return err
			}
			ui.Pass("Stopped", name)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
