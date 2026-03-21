package cmd

import (
	"context"
	"time"

	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [suite]",
	Short: "Stop all or one suite",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runStop,
}

func runStop(cmd *cobra.Command, args []string) error {
	names, err := installedSuiteTargets(args, true)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		ui.Info("Stop", "No suites installed. Run: concave install [suite]")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	for _, name := range names {
		ui.Info("Stopping", name)
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
		ui.Pass("Stopped", name)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
