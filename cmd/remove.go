package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gradientlinux/concave/internal/config"
	"github.com/gradientlinux/concave/internal/docker"
	"github.com/gradientlinux/concave/internal/suite"
	"github.com/gradientlinux/concave/internal/system"
	"github.com/gradientlinux/concave/internal/ui"
	"github.com/gradientlinux/concave/internal/workspace"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [suite]",
	Short: "Remove a suite while preserving user data",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := suite.Get(args[0])
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		composePath := workspace.ComposePath(s.Name)
		if _, err := os.Stat(composePath); err == nil {
			if err := docker.ComposeDown(ctx, composePath); err != nil {
				return err
			}
			if err := os.Remove(composePath); err != nil {
				return fmt.Errorf("remove %s: %w", composePath, err)
			}
		}

		if err := config.RemoveInstalled(s.Name); err != nil {
			return err
		}
		versions, err := config.LoadVersions()
		if err != nil {
			return err
		}
		config.RemoveSuiteVersions(versions, s.Name)
		if err := config.SaveVersions(versions); err != nil {
			return err
		}
		if err := system.Deregister(s); err != nil {
			return err
		}

		ui.Pass("Removed", s.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
