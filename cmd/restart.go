package cmd

import (
	"context"
	"time"

	"github.com/gradientlinux/concave/internal/docker"
	"github.com/gradientlinux/concave/internal/ui"
	"github.com/gradientlinux/concave/internal/workspace"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart [suite]",
	Short: "Restart a suite",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if err := docker.ComposeDown(ctx, workspace.ComposePath(args[0])); err != nil {
			return err
		}
		if err := docker.ComposeUp(ctx, workspace.ComposePath(args[0]), true); err != nil {
			return err
		}
		ui.Pass("Restarted", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
