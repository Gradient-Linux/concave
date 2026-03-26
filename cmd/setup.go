package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Gradient-Linux/concave/internal/config"
	"github.com/Gradient-Linux/concave/internal/gpu"
	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Run the first-boot concave setup wizard",
	RunE:  runSetup,
}

func runSetup(cmd *cobra.Command, args []string) error {
	return runLockedOperation("setup", 90*time.Minute, cleanupAllComposeTemps(), func(ctx context.Context) error {
		ui.Header("Gradient Linux — concave setup")
		root := workspaceRoot()
		if err := ensureWorkspaceLayout(); err != nil {
			return err
		}
		ui.Pass("Workspace", root)

		state, err := loadSetupState(root)
		if err != nil {
			return err
		}

		gpuState, err := gpuDetectState()
		if err != nil {
			return err
		}
		if isStepComplete(state, config.StepHardwareDetection) {
			ui.Info("Skipping", string(config.StepHardwareDetection)+" already complete")
		} else {
			ui.Info("GPU state", gpuState.String())
			if err := markStepComplete(root, config.StepHardwareDetection); err != nil {
				return err
			}
			state, _ = loadSetupState(root)
		}
		if isStepComplete(state, config.StepDriverInstall) {
			ui.Info("Skipping", string(config.StepDriverInstall)+" already complete")
		} else {
			if gpuState == gpu.GPUStateNVIDIA && ui.Confirm("Run NVIDIA driver verification now?") {
				if err := runGPUSetup(gpuSetupCmd, nil); err != nil {
					return err
				}
			}
			if err := markStepComplete(root, config.StepDriverInstall); err != nil {
				return err
			}
		}

		if isStepComplete(state, config.StepInternetCheck) {
			ui.Info("Skipping", string(config.StepInternetCheck)+" already complete")
		} else {
			ok, err := systemInternetReachable()
			if err != nil {
				ui.Warn("Internet", err.Error())
			} else if ok {
				ui.Pass("Internet", "reachable")
			}
			if err := markStepComplete(root, config.StepInternetCheck); err != nil {
				return err
			}
		}

		if err := ensureDockerRuntime(ctx, "setup"); err != nil {
			return err
		}

		selected := []string{"boosting"}
		if isStepComplete(state, config.StepSuiteSelection) {
			ui.Info("Skipping", string(config.StepSuiteSelection)+" already complete")
			current, err := loadState()
			if err == nil && len(current.Installed) > 0 {
				selected = orderInstalledSuites(current.Installed, false)
			}
		} else {
			selected = ui.Checklist(suiteNames())
			if len(selected) == 0 {
				selected = []string{"boosting"}
			}
			if err := markStepComplete(root, config.StepSuiteSelection); err != nil {
				return err
			}
		}

		if isStepComplete(state, config.StepImagePull) {
			ui.Info("Skipping", string(config.StepImagePull)+" already complete")
		} else {
			for _, name := range selected {
				if name == "forge" {
					continue
				}
				installed, err := isInstalled(name)
				if err != nil {
					return err
				}
				if installed {
					ui.Info("Skipping", name+" already installed")
					continue
				}
				if err := installSuite(ctx, name, suite.InstallOptions{GPUAvailable: gpuState == gpu.GPUStateNVIDIA}); err != nil {
					return fmt.Errorf("setup install %s: %w", name, err)
				}
			}
			if err := markStepComplete(root, config.StepImagePull); err != nil {
				return err
			}
		}

		if isStepComplete(state, config.StepHealthCheck) {
			ui.Info("Skipping", string(config.StepHealthCheck)+" already complete")
		} else {
			for _, name := range selected {
				installed, err := isInstalled(name)
				if err != nil {
					return err
				}
				if !installed {
					continue
				}
				s, err := currentSuiteDefinition(name)
				if err != nil {
					return err
				}
				if err := dockerComposeUp(ctx, dockerComposePath(name), true); err != nil {
					return wrapDockerError(err)
				}
				if err := systemRegisterPorts(s); err != nil {
					return err
				}
				if err := waitForHealthy(ctx, s); err != nil {
					return err
				}
			}
			if err := doctorCmd.RunE(doctorCmd, nil); err != nil {
				return err
			}
			if err := markStepComplete(root, config.StepHealthCheck); err != nil {
				return err
			}
		}

		if err := markSetupComplete(root); err != nil {
			return err
		}
		markerPath := filepath.Join(os.Getenv("HOME"), ".config", "concave", "setup-complete")
		if err := os.MkdirAll(filepath.Dir(markerPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(markerPath, []byte(time.Now().UTC().Format(time.RFC3339)), 0o644); err != nil {
			return err
		}
		ui.Pass("Setup", "complete")
		return nil
	})
}

func cleanupAllComposeTemps() func() {
	return func() {
		for _, name := range suiteNames() {
			_ = os.Remove(dockerComposePath(name) + ".tmp")
		}
	}
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
