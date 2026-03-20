package cmd

import (
	"fmt"

	"github.com/gradient-linux/concave/internal/gpu"
	"github.com/gradient-linux/concave/internal/suite"
	"github.com/gradient-linux/concave/internal/ui"
	"github.com/gradient-linux/concave/internal/workspace"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Run the first-boot concave setup wizard",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Header("Gradient Linux — concave setup")
		if err := workspace.EnsureLayout(); err != nil {
			return err
		}
		ui.Pass("Workspace", workspace.Root())

		state, err := gpu.Detect()
		if err != nil {
			return err
		}
		ui.Info("GPU state", state.String())
		if state == gpu.GPUStateNVIDIA && ui.Confirm("Run NVIDIA driver verification now?") {
			if err := driverWizardCmd.RunE(driverWizardCmd, nil); err != nil {
				return err
			}
		}

		selected := ui.Checklist(suite.Names())
		if len(selected) == 0 {
			selected = []string{"boosting"}
		}
		for _, name := range selected {
			if name == "forge" {
				continue
			}
			if err := installCmd.RunE(installCmd, []string{name}); err != nil {
				return fmt.Errorf("setup install %s: %w", name, err)
			}
		}
		if err := doctorCmd.RunE(doctorCmd, nil); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
