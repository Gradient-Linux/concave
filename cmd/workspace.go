package cmd

import (
	"fmt"

	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var workspacePruneOutputs bool

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage the Gradient workspace",
}

var workspaceInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create the ~/gradient workspace tree",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureWorkspaceLayout(); err != nil {
			return fmt.Errorf("workspace init: %w", err)
		}
		ui.Pass("Workspace", workspaceRoot())
		return nil
	},
}

var workspaceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show workspace disk usage",
	RunE: func(cmd *cobra.Command, args []string) error {
		usages, err := workspaceStatus()
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
		path, err := workspaceBackup()
		if err != nil {
			return fmt.Errorf("workspace backup: %w", err)
		}
		ui.Pass("Backup", path)
		return nil
	},
}

var workspacePruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Prune generated workspace directories",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !workspacePruneOutputs {
			return fmt.Errorf("workspace prune requires --outputs")
		}
		if err := workspaceClean(); err != nil {
			return fmt.Errorf("workspace prune: %w", err)
		}
		ui.Pass("Outputs", "cleaned")
		return nil
	},
}

var workspaceCleanCmd = &cobra.Command{
	Use:        "clean",
	Short:      "Deprecated: use 'concave workspace prune'",
	Deprecated: "use 'concave workspace prune' instead",
	Hidden:     true,
	RunE:       workspacePruneCmd.RunE,
}

func init() {
	workspacePruneCmd.Flags().BoolVar(&workspacePruneOutputs, "outputs", false, "prune ~/gradient/outputs contents")
	workspaceCleanCmd.Flags().BoolVar(&workspacePruneOutputs, "outputs", false, "prune ~/gradient/outputs contents")
	workspaceCmd.AddCommand(workspaceInitCmd, workspaceStatusCmd, workspaceBackupCmd, workspacePruneCmd, workspaceCleanCmd)
	rootCmd.AddCommand(workspaceCmd)
}
