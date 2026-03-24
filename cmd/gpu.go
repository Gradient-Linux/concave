package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/Gradient-Linux/concave/internal/gpu"
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var gpuCmd = &cobra.Command{
	Use:   "gpu",
	Short: "GPU driver management and health checks",
}

var gpuSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install or repair GPU drivers",
	Long:  "Interactive GPU driver installer. Detects GPU, recommends driver branch, installs CUDA stack.",
	RunE:  runGPUSetup,
}

var gpuCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Run GPU health checks",
	RunE:  runGPUCheck,
}

var gpuInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show GPU details",
	RunE:  runGPUInfo,
}

var driverWizardCmd = &cobra.Command{
	Use:        "driver-wizard",
	Short:      "Deprecated: use 'concave gpu setup'",
	Deprecated: "use 'concave gpu setup' instead",
	Hidden:     true,
	RunE:       runGPUSetup,
}

func runDriverWizard(cmd *cobra.Command, args []string) error {
	return runGPUSetup(cmd, args)
}

func runGPUSetup(cmd *cobra.Command, args []string) error {
	return runLockedOperation("gpu setup", 5*time.Minute, nil, func(ctx context.Context) error {
		_ = ctx
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
	})
}

func runGPUCheck(cmd *cobra.Command, args []string) error {
	ui.Header("Gradient Linux — concave gpu check")
	runGPUCheckSummary()

	if state, err := gpuDetectState(); err == nil && state == gpu.GPUStateNVIDIA {
		if ok, toolkitErr := gpuToolkitConfigured(); toolkitErr != nil {
			ui.Warn("Toolkit", toolkitErr.Error())
		} else if ok {
			ui.Pass("Toolkit", "nvidia-container-toolkit configured")
		}
	}
	return nil
}

func runGPUInfo(cmd *cobra.Command, args []string) error {
	ui.Header("Gradient Linux — concave gpu info")

	state, err := gpuDetectState()
	if err != nil {
		return err
	}
	switch state {
	case gpu.GPUStateNone:
		ui.Info("GPU", "cpu-only")
		return nil
	case gpu.GPUStateAMD:
		ui.Warn("GPU", "AMD detected — ROCm support coming in Gradient Linux v0.3")
		return nil
	}

	devices, err := gpuNVIDIADevices()
	if err != nil {
		return err
	}
	capability, capErr := gpuComputeCapability()
	branch, branchErr := gpuRecommendedDriverBranch()
	for _, device := range devices {
		label := fmt.Sprintf("GPU %d", device.Index)
		detail := fmt.Sprintf("%s · %d MiB / %d MiB used · driver %s", device.Name, device.MemoryUsedMiB, device.MemoryTotalMiB, device.DriverVersion)
		ui.Info(label, detail)
	}
	if capErr == nil {
		ui.Info("Compute cap", capability)
	} else {
		ui.Warn("Compute cap", capErr.Error())
	}
	if branchErr == nil {
		ui.Info("Driver branch", branch)
	} else {
		ui.Warn("Driver branch", branchErr.Error())
	}
	return nil
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
	ui.Info("Option B", "exit — disable Secure Boot in BIOS, then rerun concave gpu setup")
	if !uiConfirm("Continue with MOK enrollment guidance?") {
		ui.Warn("Secure Boot", "exit selected — disable Secure Boot in BIOS and rerun concave gpu setup")
		return false, nil
	}
	ui.Info("Secure Boot", "continue — enroll the MOK key on next reboot")
	return true, nil
}

func init() {
	gpuCmd.AddCommand(gpuSetupCmd, gpuCheckCmd, gpuInfoCmd)
	rootCmd.AddCommand(gpuCmd)
	rootCmd.AddCommand(driverWizardCmd)
}
