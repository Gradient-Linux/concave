package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [suite]",
	Short: "Remove a suite while preserving user data",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
	name := args[0]
	installed, err := isInstalled(name)
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("suite %s is not installed", name)
	}

	if !uiConfirm("This will remove " + name + " containers and images. Your data in ~/gradient/ will not be touched. Continue? [y/N]") {
		return nil
	}

	return runLockedOperation("remove", 5*time.Minute, composeCleanup(name), func(ctx context.Context) error {
		composePath := dockerComposePath(name)
		if _, statErr := os.Stat(composePath); statErr == nil {
			ui.Progress("Remove", 0, 4)
			if err := runDockerInteractive(ctx, "compose", "-f", composePath, "down", "--rmi", "all"); err != nil {
				return wrapDockerError(fmt.Errorf("docker compose down %s: %w", composePath, err))
			}
		} else if os.IsNotExist(statErr) {
			ui.Warn("Remove", fmt.Sprintf("%s missing — cleaning suite state directly", composePath))
			if err := cleanupSuiteContainers(ctx, name); err != nil {
				return err
			}
		} else {
			return wrapDockerError(fmt.Errorf("stat %s: %w", composePath, statErr))
		}
		ui.Progress("Remove", 1, 4)
		_ = os.Remove(composePath)

		if err := removeSuite(name); err != nil {
			return err
		}
		ui.Progress("Remove", 2, 4)

		manifest, err := loadManifest()
		if err != nil {
			return err
		}
		delete(manifest, name)
		if err := saveManifest(manifest); err != nil {
			return err
		}

		s, err := cleanupSuiteDefinition(name)
		if err != nil {
			return err
		}
		ui.Progress("Remove", 3, 4)
		if err := systemDeregisterPorts(s); err != nil {
			return err
		}

		ui.Progress("Remove", 4, 4)
		ui.Pass("Removed", name)
		return nil
	})
}

func cleanupSuiteDefinition(name string) (suite.Suite, error) {
	s, err := currentSuiteDefinition(name)
	if err == nil {
		return s, nil
	}
	if name == "forge" && strings.Contains(err.Error(), "forge has no recorded component selection") {
		return getSuite(name)
	}
	return suite.Suite{}, err
}

func cleanupSuiteContainers(ctx context.Context, name string) error {
	s, err := cleanupSuiteDefinition(name)
	if err != nil {
		return err
	}
	names := containerNames(s)
	if len(names) == 0 {
		return nil
	}
	args := append([]string{"rm", "-f"}, names...)
	if _, err := runDockerOutput(ctx, args...); err != nil {
		text := strings.ToLower(err.Error())
		if !strings.Contains(text, "no such") {
			return wrapDockerError(fmt.Errorf("docker %s: %w", strings.Join(args, " "), err))
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
