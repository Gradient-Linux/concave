package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [suite]",
	Short: "Update a suite to the images pinned in the registry",
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	name := args[0]
	installed, err := isInstalled(name)
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("suite %s is not installed", name)
	}

	s, err := currentSuiteDefinition(name)
	if err != nil {
		return err
	}

	manifest, err := loadManifest()
	if err != nil {
		return err
	}

	return runLockedOperation("update", 5*time.Minute, composeCleanup(name), func(ctx context.Context) error {
		totalSteps := len(s.Containers) + 4
		step := 0
		ui.Progress("Update", step, totalSteps)

		for _, container := range s.Containers {
			current := ""
			if containers, ok := manifest[s.Name]; ok {
				current = containers[container.Name].Current
			}
			ui.Info("Pulling", container.Image)
			if err := dockerPullWithRollbackSafety(ctx, container.Image, nil); err != nil {
				return wrapDockerError(err)
			}
			manifest = recordUpdate(manifest, s.Name, container.Name, container.Image)
			ui.Info(container.Name, fmt.Sprintf("%s -> %s", current, container.Image))
			step++
			ui.Progress("Update", step, totalSteps)
		}

		if err := saveManifest(manifest); err != nil {
			return err
		}
		step++
		ui.Progress("Update", step, totalSteps)
		if _, err := writeComposeForCurrentState(name); err != nil {
			return wrapDockerError(err)
		}
		step++
		ui.Progress("Update", step, totalSteps)
		if err := dockerComposeUp(ctx, dockerComposePath(name), true); err != nil {
			return wrapDockerError(err)
		}
		step++
		ui.Progress("Update", step, totalSteps)
		if err := waitForHealthy(ctx, s); err != nil {
			return err
		}
		ui.Progress("Update", totalSteps, totalSteps)

		ui.Pass("Update", name)
		return nil
	})
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
