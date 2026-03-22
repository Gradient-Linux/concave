package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback [suite]",
	Short: "Rollback a suite to the previous recorded image tags",
	Args:  cobra.ExactArgs(1),
	RunE:  runRollback,
}

func runRollback(cmd *cobra.Command, args []string) error {
	name := args[0]
	installed, err := isInstalled(name)
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("suite %s is not installed", name)
	}

	manifest, err := loadManifest()
	if err != nil {
		return err
	}
	manifest, err = swapForRollback(manifest, name)
	if err != nil {
		if strings.Contains(err.Error(), "no previous version") {
			return fmt.Errorf("Nothing to roll back — run concave update first")
		}
		return err
	}
	if err := saveManifest(manifest); err != nil {
		return err
	}

	return runLockedOperation("rollback", 5*time.Minute, composeCleanup(name), func(ctx context.Context) error {
		totalSteps := 4
		ui.Progress("Rollback", 0, totalSteps)
		if _, err := writeComposeForCurrentState(name); err != nil {
			return wrapDockerError(err)
		}
		ui.Progress("Rollback", 1, totalSteps)
		if err := dockerComposeDown(ctx, dockerComposePath(name)); err != nil {
			return wrapDockerError(err)
		}
		ui.Progress("Rollback", 2, totalSteps)
		if err := dockerComposeUp(ctx, dockerComposePath(name), true); err != nil {
			return wrapDockerError(err)
		}

		s, err := currentSuiteDefinition(name)
		if err != nil {
			return err
		}
		ui.Progress("Rollback", 3, totalSteps)
		if err := waitForHealthy(ctx, s); err != nil {
			return err
		}

		ui.Progress("Rollback", totalSteps, totalSteps)
		ui.Pass("Rollback", name)
		return nil
	})
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}
