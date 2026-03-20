package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/gradient-linux/concave/internal/suite"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell [suite]",
	Short: "Open an interactive shell in a suite's primary container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := suite.Get(args[0])
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		container := suite.PrimaryContainer(s)

		command := exec.CommandContext(ctx, "docker", "exec", "-it", container, "bash")
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		if err := command.Run(); err == nil {
			return nil
		}

		command = exec.CommandContext(ctx, "docker", "exec", "-it", container, "sh")
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		if err := command.Run(); err != nil {
			return fmt.Errorf("docker exec shell %s: %w", container, err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
}
