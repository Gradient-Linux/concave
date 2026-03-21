package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/Gradient-Linux/concave/internal/ui"
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
		composePath := workspaceComposePath(s.Name)
		if _, err := os.Stat(composePath); err == nil {
			if err := dockerComposeDown(ctx, composePath); err != nil {
				return err
			}
			if err := os.Remove(composePath); err != nil {
				return fmt.Errorf("remove %s: %w", composePath, err)
			}
		}

		if err := removeInstalledSuite(s.Name); err != nil {
			return err
		}
		versions, err := loadVersions()
		if err != nil {
			return err
		}
		removeSuiteVersions(versions, s.Name)
		if err := saveVersions(versions); err != nil {
			return err
		}
		if err := systemDeregisterPorts(s); err != nil {
			return err
		}

		ui.Pass("Removed", s.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
