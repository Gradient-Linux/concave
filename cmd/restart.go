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

	return runLockedOperation("restart", 5*time.Minute, nil, func(ctx context.Context) error {
		name := names[0]
		totalSteps := 4
		step := 0
		ui.Info("Restarting", name)
		if err := dockerComposeDown(ctx, dockerComposePath(name)); err != nil {
			return wrapDockerError(err)
		}
		step++
		ui.Progress("Restart", step, totalSteps)
		s, err := currentSuiteDefinition(name)
		if err != nil {
			return err
		}
		if err := systemDeregisterPorts(s); err != nil {
			return err
		}
		if err := dockerComposeUp(ctx, dockerComposePath(name), true); err != nil {
			return wrapDockerError(err)
		}
		step++
		ui.Progress("Restart", step, totalSteps)
		if err := systemRegisterPorts(s); err != nil {
			return err
		}
		if err := waitForHealthy(ctx, s); err != nil {
			return err
		}
		ui.Progress("Restart", totalSteps, totalSteps)
		ui.Pass("Restarted", name)
		return nil
	})
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
