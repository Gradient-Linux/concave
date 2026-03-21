package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell [suite]",
	Short: "Open an interactive shell in a suite's primary container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getSuite(args[0])
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		container := primaryContainer(s)

		if err := runDockerInteractive(ctx, "exec", "-it", container, "bash"); err == nil {
			return nil
		}

		if err := runDockerInteractive(ctx, "exec", "-it", container, "sh"); err != nil {
			return fmt.Errorf("docker exec shell %s: %w", container, err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
}
