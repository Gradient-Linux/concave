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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	for _, container := range s.Containers {
		current := ""
		if containers, ok := manifest[s.Name]; ok {
			current = containers[container.Name].Current
		}
		ui.Info("Pulling", container.Image)
		if err := dockerPullWithRollbackSafety(ctx, container.Image, nil); err != nil {
			return err
		}
		manifest = recordUpdate(manifest, s.Name, container.Name, container.Image)
		ui.Info(container.Name, fmt.Sprintf("%s -> %s", current, container.Image))
	}

	if err := saveManifest(manifest); err != nil {
		return err
	}
	if _, err := writeComposeForCurrentState(name); err != nil {
		return err
	}
	if err := dockerComposeUp(ctx, dockerComposePath(name), true); err != nil {
		return err
	}

	ui.Pass("Update", name)
	return nil
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
