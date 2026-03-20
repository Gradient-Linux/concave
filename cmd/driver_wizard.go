package cmd

import (
	"fmt"

	"github.com/gradient-linux/concave/internal/gpu"
	"github.com/gradient-linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var driverWizardCmd = &cobra.Command{
	Use:   "driver-wizard",
	Short: "Guide NVIDIA driver and toolkit verification",
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := gpu.Detect()
		if err != nil {
			return err
		}

		switch state {
		case gpu.GPUStateNone:
			ui.Warn("GPU", "CPU-only host detected — no driver changes required")
			return nil
		case gpu.GPUStateAMD:
			gpu.DetectAMD()
			return nil
		default:
			branch, err := gpu.RecommendedDriverBranch()
			if err != nil {
				return err
			}
			ui.Info("Recommended driver", branch)
			if ok, err := gpu.ToolkitConfigured(); err != nil {
				ui.Warn("Toolkit", err.Error())
			} else if ok {
				ui.Pass("Toolkit", "nvidia-container-toolkit configured")
			}
			if err := gpu.VerifyPassthrough(); err != nil {
				return fmt.Errorf("gpu passthrough verification: %w", err)
			}
			ui.Pass("GPU", "docker passthrough verified")
			return nil
		}
	},
}

func init() {
	rootCmd.AddCommand(driverWizardCmd)
}
