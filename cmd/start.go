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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	for _, name := range names {
		ui.Info("Starting", name)
		if err := dockerComposeUp(ctx, dockerComposePath(name), true); err != nil {
			return err
		}
		s, err := currentSuiteDefinition(name)
		if err != nil {
			return err
		}
		if err := systemRegisterPorts(s); err != nil {
			return err
		}
		ui.Pass("Started", name)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(startCmd)
}
