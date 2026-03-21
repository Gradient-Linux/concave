package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	composePath := dockerComposePath(name)
	if err := runDockerInteractive(ctx, "compose", "-f", composePath, "down", "--rmi", "all"); err != nil {
		return fmt.Errorf("docker compose down %s: %w", composePath, err)
	}
	_ = os.Remove(composePath)

	if err := removeSuite(name); err != nil {
		return err
	}

	manifest, err := loadManifest()
	if err != nil {
		return err
	}
	delete(manifest, name)
	if err := saveManifest(manifest); err != nil {
		return err
	}

	s, err := getSuite(name)
	if err != nil {
		return err
	}
	if err := systemDeregisterPorts(s); err != nil {
		return err
	}

	ui.Pass("Removed", name)
	return nil
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
