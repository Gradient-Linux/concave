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
	RunE:  runDriverWizard,
}

func runDriverWizard(cmd *cobra.Command, args []string) error {
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
	case gpu.GPUStateNVIDIA:
		return runNVIDIAWizard()
	default:
		return fmt.Errorf("unknown GPU state")
	}
}

func runNVIDIAWizard() error {
	shouldContinue, err := confirmSecureBootFlow()
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
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

func confirmSecureBootFlow() (bool, error) {
	secureBootEnabled, err := gpuSecureBootEnabled()
	if err != nil {
		ui.Warn("Secure Boot", err.Error())
		return true, nil
	}
	if !secureBootEnabled {
		ui.Pass("Secure Boot", "disabled")
		return true, nil
	}

	ui.Warn("Secure Boot", "enabled")
	ui.Info("Option A", "continue — enroll a MOK key on next reboot")
	ui.Info("Option B", "exit — disable Secure Boot in BIOS, then rerun concave driver-wizard")
	if !uiConfirm("Continue with MOK enrollment guidance?") {
		ui.Warn("Secure Boot", "exit selected — disable Secure Boot in BIOS and rerun concave driver-wizard")
		return false, nil
	}
	ui.Info("Secure Boot", "continue — enroll the MOK key on next reboot")
	return true, nil
}

func init() {
	rootCmd.AddCommand(driverWizardCmd)
}
