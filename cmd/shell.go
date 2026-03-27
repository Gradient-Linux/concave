package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell [suite]",
	Short: "Open an interactive shell in a suite's primary container",
	Args:  cobra.ExactArgs(1),
	RunE:  runShell,
}

func runShell(cmd *cobra.Command, args []string) error {
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
	container := primaryContainer(s)
	composePath := dockerComposePath(name)
	status, err := dockerComposeServiceStatus(context.Background(), composePath, container)
	if err != nil {
		return err
	}
	if status != "running" {
		return fmt.Errorf("suite %s is not running. Run: concave start %s", name, name)
	}

	if err := dockerComposeExecInteractive(context.Background(), composePath, container, "bash"); err == nil {
		return nil
	}
	if err := dockerComposeExecInteractive(context.Background(), composePath, container, "sh"); err != nil {
		return fmt.Errorf("docker exec shell %s: %w", container, err)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(shellCmd)
}
