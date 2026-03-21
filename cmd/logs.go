package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var logsService string

var logsCmd = &cobra.Command{
	Use:   "logs [suite]",
	Short: "Tail suite logs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		command := []string{"compose", "-f", workspaceComposePath(args[0]), "logs", "-f", "--tail=100"}
		if logsService != "" {
			command = append(command, logsService)
		}

		if err := runDockerInteractive(ctx, command...); err != nil {
			return fmt.Errorf("docker compose logs %s: %w", args[0], err)
		}
		return nil
	},
}

func init() {
	logsCmd.Flags().StringVar(&logsService, "service", "", "single service to tail")
	rootCmd.AddCommand(logsCmd)
}
