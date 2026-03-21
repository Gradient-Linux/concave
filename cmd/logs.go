package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	logsService string
	logsLines   int
	logsFollow  bool
)

var logsCmd = &cobra.Command{
	Use:   "logs [suite]",
	Short: "Tail logs from a suite's containers",
	Args:  cobra.ExactArgs(1),
	RunE:  runLogs,
}

func runLogs(cmd *cobra.Command, args []string) error {
	name := args[0]
	installed, err := isInstalled(name)
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("suite %s is not installed", name)
	}

	command := []string{"compose", "-f", dockerComposePath(name), "logs"}
	if logsFollow {
		command = append(command, "--follow")
	}
	command = append(command, "--tail", strconv.Itoa(logsLines))
	if logsService != "" {
		command = append(command, logsService)
	}

	if err := runDockerInteractive(context.Background(), command...); err != nil {
		return fmt.Errorf("docker compose logs %s: %w", name, err)
	}
	return nil
}

func init() {
	logsCmd.Flags().StringVar(&logsService, "service", "", "single service to tail")
	logsCmd.Flags().IntVar(&logsLines, "lines", 50, "number of historical log lines to show")
	logsCmd.Flags().BoolVar(&logsFollow, "follow", true, "follow log output")
	rootCmd.AddCommand(logsCmd)
}
