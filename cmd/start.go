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
	RunE: func(cmd *cobra.Command, args []string) error {
		names, err := targetSuites(args)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		for _, name := range names {
			if err := dockerComposeUp(ctx, workspaceComposePath(name), true); err != nil {
				return err
			}
			ui.Pass("Started", name)
		}
		return nil
	},
}

func targetSuites(args []string) ([]string, error) {
	if len(args) == 1 {
		if _, err := getSuite(args[0]); err != nil {
			return nil, err
		}
		return []string{args[0]}, nil
	}
	state, err := loadState()
	if err != nil {
		return nil, err
	}
	return state.Installed, nil
}

func init() {
	rootCmd.AddCommand(startCmd)
}
