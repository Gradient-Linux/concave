package cmd

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec [suite] -- [command]",
	Short: "Run a non-interactive command inside a suite container",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runExec,
}

func runExec(cmd *cobra.Command, args []string) error {
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

	if err := dockerComposeExecInteractive(context.Background(), composePath, container, args[1:]...); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr
		}
		if _, ok := resolveExitCode(err); ok {
			return err
		}
		return fmt.Errorf("suite exec %s: %w", name, err)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(execCmd)
}
