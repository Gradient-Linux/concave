package cmd

import (
	"fmt"

	"github.com/Gradient-Linux/concave/internal/gpu"
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var driverWizardCmd = &cobra.Command{
	Use:   "driver-wizard",
	Short: "Guide NVIDIA driver and toolkit verification",
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := gpuDetectState()
		if err != nil {
			return err
		}

		switch state {
		case gpu.GPUStateNone:
			ui.Warn("GPU", "CPU-only host detected — no driver changes required")
			return nil
		case gpu.GPUStateAMD:
			gpuDetectAMDState()
			return nil
		default:
			secureBootEnabled, err := gpuSecureBootEnabled()
			if err != nil {
				ui.Warn("Secure Boot", err.Error())
			} else if secureBootEnabled {
				ui.Warn("Secure Boot", "enabled")
				if !ui.Confirm("Continue and enroll a MOK key on next reboot?") {
					ui.Warn("Secure Boot", "disable Secure Boot in BIOS and rerun concave driver-wizard")
					return nil
				}
				ui.Info("Secure Boot", "continue — enroll the MOK key on next reboot")
			}

			branch, err := gpuRecommendedDriverBranch()
			if err != nil {
				return err
			}
			ui.Info("Recommended driver", branch)
			if ok, err := gpuToolkitConfigured(); err != nil {
				ui.Warn("Toolkit", err.Error())
			} else if ok {
				ui.Pass("Toolkit", "nvidia-container-toolkit configured")
			}
			if err := gpuVerifyPassthrough(); err != nil {
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
