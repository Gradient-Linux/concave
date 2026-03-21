package cmd

import (
	"context"
	"time"

	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/Gradient-Linux/concave/internal/ui"
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

		versions, err := loadVersions()
		if err != nil {
			return err
		}
		if err := swapPreviousVersions(versions, s.Name); err != nil {
			return err
		}
		if err := saveVersions(versions); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if _, err := dockerWriteSuiteCompose(ctx, s); err != nil {
			return err
		}
		if err := dockerComposeDown(ctx, workspaceComposePath(s.Name)); err != nil {
			return err
		}
		if err := dockerComposeUp(ctx, workspaceComposePath(s.Name), true); err != nil {
			return err
		}
		ui.Pass("Rollback", s.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}
