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

	return runLockedOperation("update", 45*time.Minute, composeCleanup(name), func(ctx context.Context) error {
		totalSteps := len(s.Containers) + 4
		step := 0
		ui.Progress("Preparing update", step, totalSteps)

		for _, container := range s.Containers {
			current := ""
			if containers, ok := manifest[s.Name]; ok {
				current = containers[container.Name].Current
			}
			ui.Info("Pulling", container.Image)
			label := progressImageLabel(container.Image)
			ui.Progress("Pulling "+label, step, totalSteps)
			if err := dockerPullWithRollbackSafety(ctx, container.Image, ui.DockerPullReporter("Pulling "+label)); err != nil {
				return wrapDockerError(err)
			}
			manifest = recordUpdate(manifest, s.Name, container.Name, container.Image)
			ui.Info(container.Name, fmt.Sprintf("%s -> %s", current, container.Image))
			step++
			ui.Progress("Pulled "+label, step, totalSteps)
		}

		if err := saveManifest(manifest); err != nil {
			return err
		}
		step++
		ui.Progress("Saving version manifest", step, totalSteps)
		if _, err := writeComposeForCurrentState(name); err != nil {
			return wrapDockerError(err)
		}
		step++
		ui.Progress("Writing compose file", step, totalSteps)
		if err := dockerComposeUp(ctx, dockerComposePath(name), true); err != nil {
			return wrapDockerError(err)
		}
		step++
		ui.Progress("Starting updated services", step, totalSteps)
		if err := waitForHealthy(ctx, s); err != nil {
			return err
		}
		ui.Progress("Update complete", totalSteps, totalSteps)

		ui.Pass("Update", name)
		return nil
	})
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
