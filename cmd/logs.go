package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/gradient-linux/concave/internal/workspace"
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

		command := []string{"compose", "-f", workspace.ComposePath(args[0]), "logs", "-f", "--tail=100"}
		if logsService != "" {
			command = append(command, logsService)
		}

		execCmd := exec.CommandContext(ctx, "docker", command...)
		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		if err := execCmd.Run(); err != nil {
			return fmt.Errorf("docker compose logs %s: %w", args[0], err)
		}
		return nil
	},
}

func init() {
	logsCmd.Flags().StringVar(&logsService, "service", "", "single service to tail")
	rootCmd.AddCommand(logsCmd)
}
