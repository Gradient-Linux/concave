package cmd

import (
	"context"
	"time"

	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart [suite]",
	Short: "Restart a suite",
	Args:  cobra.ExactArgs(1),
	RunE:  runRestart,
}

func runRestart(cmd *cobra.Command, args []string) error {
	names, err := installedSuiteTargets(args, false)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	name := names[0]
	ui.Info("Restarting", name)
	if err := dockerComposeDown(ctx, dockerComposePath(name)); err != nil {
		return err
	}
	s, err := currentSuiteDefinition(name)
	if err != nil {
		return err
	}
	if err := systemDeregisterPorts(s); err != nil {
		return err
	}
	if err := dockerComposeUp(ctx, dockerComposePath(name), true); err != nil {
		return err
	}
	if err := systemRegisterPorts(s); err != nil {
		return err
	}
	ui.Pass("Restarted", name)
	return nil
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
