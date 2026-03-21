package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec [suite] -- [command]",
	Short: "Run a non-interactive command inside a suite container",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getSuite(args[0])
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := dockerExecCommand(ctx, primaryContainer(s), args[1:]...); err != nil {
			return fmt.Errorf("suite exec %s: %w", s.Name, err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}
