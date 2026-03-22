package cmd

import (
	"context"
	"time"

	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [suite]",
	Short: "Start all or one suite",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runStart,
}

func runStart(cmd *cobra.Command, args []string) error {
	names, err := installedSuiteTargets(args, false)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		ui.Info("Start", "No suites installed. Run: concave install [suite]")
		return nil
	}

	return runLockedOperation("start", 5*time.Minute, nil, func(ctx context.Context) error {
		totalSteps := len(names) * 2
		if totalSteps == 0 {
			totalSteps = 1
		}
		step := 0
		for _, name := range names {
			ui.Info("Starting", name)
			if err := dockerComposeUp(ctx, dockerComposePath(name), true); err != nil {
				return wrapDockerError(err)
			}
			step++
			ui.Progress("Start", step, totalSteps)
			s, err := currentSuiteDefinition(name)
			if err != nil {
				return err
			}
			if err := systemRegisterPorts(s); err != nil {
				return err
			}
			if err := waitForHealthy(ctx, s); err != nil {
				return err
			}
			step++
			ui.Progress("Start", step, totalSteps)
			ui.Pass("Started", name)
		}
		return nil
	})
}

func init() {
	rootCmd.AddCommand(startCmd)
}
