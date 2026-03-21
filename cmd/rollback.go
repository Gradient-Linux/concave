package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Gradient-Linux/concave/internal/suite"
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if _, err := writeComposeForCurrentState(name); err != nil {
		return err
	}
	if err := dockerComposeDown(ctx, dockerComposePath(name)); err != nil {
		return err
	}
	if err := dockerComposeUp(ctx, dockerComposePath(name), true); err != nil {
		return err
	}

	s, err := currentSuiteDefinition(name)
	if err != nil {
		return err
	}
	if err := waitForRunning(ctx, s); err != nil {
		return err
	}

	ui.Pass("Rollback", name)
	return nil
}

func waitForRunning(parent context.Context, s suite.Suite) error {
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	containerList := containerNames(s)
	for {
		allRunning := true
		for _, container := range containerList {
			status, err := dockerContainerStatus(ctx, container)
			if err != nil || status != "running" {
				allRunning = false
				break
			}
		}
		if allRunning {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for %s to become healthy", strings.Join(containerList, ", "))
		case <-ticker.C:
		}
	}
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}
