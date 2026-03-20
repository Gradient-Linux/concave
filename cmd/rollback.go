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

var rollbackCmd = &cobra.Command{
	Use:   "rollback [suite]",
	Short: "Rollback a suite to the previous recorded image tags",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := suite.Get(args[0])
		if err != nil {
			return err
		}

		versions, err := config.LoadVersions()
		if err != nil {
			return err
		}
		if err := config.SwapPrevious(versions, s.Name); err != nil {
			return err
		}
		if err := config.SaveVersions(versions); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if _, err := docker.WriteSuiteCompose(ctx, s); err != nil {
			return err
		}
		if err := docker.ComposeDown(ctx, workspace.ComposePath(s.Name)); err != nil {
			return err
		}
		if err := docker.ComposeUp(ctx, workspace.ComposePath(s.Name), true); err != nil {
			return err
		}
		ui.Pass("Rollback", s.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}
