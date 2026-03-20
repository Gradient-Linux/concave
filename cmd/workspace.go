package cmd

import (
	"fmt"

	"github.com/gradientlinux/concave/internal/ui"
	"github.com/gradientlinux/concave/internal/workspace"
	"github.com/spf13/cobra"
)

var workspaceCleanOutputs bool

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage the Gradient workspace",
}

var workspaceInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create the ~/gradient workspace tree",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := workspace.EnsureLayout(); err != nil {
			return fmt.Errorf("workspace init: %w", err)
		}
		ui.Pass("Workspace", workspace.Root())
		return nil
	},
}

var workspaceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show workspace disk usage",
	RunE: func(cmd *cobra.Command, args []string) error {
		usages, err := workspace.Status()
		if err != nil {
			return fmt.Errorf("workspace status: %w", err)
		}

		ui.Header("Gradient Linux — workspace status")
		for _, usage := range usages {
			ui.Info(usage.Name, usage.Human())
		}

		return nil
	},
}

var workspaceBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup notebooks and models into ~/gradient/backups",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := workspace.Backup()
		if err != nil {
			return fmt.Errorf("workspace backup: %w", err)
		}
		ui.Pass("Backup", path)
		return nil
	},
}

var workspaceCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean generated workspace directories",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !workspaceCleanOutputs {
			return fmt.Errorf("workspace clean requires --outputs")
		}
		if err := workspace.CleanOutputs(); err != nil {
			return fmt.Errorf("workspace clean: %w", err)
		}
		ui.Pass("Outputs", "cleaned")
		return nil
	},
}

func init() {
	workspaceCleanCmd.Flags().BoolVar(&workspaceCleanOutputs, "outputs", false, "clean ~/gradient/outputs contents")
	workspaceCmd.AddCommand(workspaceInitCmd, workspaceStatusCmd, workspaceBackupCmd, workspaceCleanCmd)
	rootCmd.AddCommand(workspaceCmd)
}
