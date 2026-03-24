package cmd

import (
	"fmt"
	"os"

	"github.com/Gradient-Linux/concave/internal/gpu"
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:        "doctor",
	Short:      "Deprecated: use 'concave check'",
	Deprecated: "use 'concave check' instead",
	Hidden:     true,
	RunE:       runCheck,
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run concave system health checks",
	RunE:  runCheck,
}

func runCheck(cmd *cobra.Command, args []string) error {
	ui.Header("Gradient Linux — concave check")

	if ok, err := systemDockerRunning(); err != nil {
		ui.Fail("Docker", err.Error())
	} else if ok {
		ui.Pass("Docker", "running")
	} else {
		ui.Fail("Docker", "not running")
	}

	if ok, err := systemUserInDockerGroup(); err != nil {
		ui.Fail("Docker group", err.Error())
	} else if ok {
		ui.Pass("Docker group", "membership detected")
	} else {
		ui.Warn("Docker group", "user not in docker group")
	}

	if ok, err := systemInternetReachable(); err != nil {
		ui.Fail("Internet", err.Error())
	} else if ok {
		ui.Pass("Internet", "reachable")
	} else {
		ui.Warn("Internet", "not reachable")
	}

	if workspaceExists() {
		ui.Pass("Workspace", workspaceRoot())
	} else {
		ui.Warn("Workspace", "not initialized")
	}

	// GPU_SECTION_START — GPU Agent adds checks here in Phase 4
	runGPUCheckSummary()
	// GPU_SECTION_END

	// RESOLVER_SECTION_START — Phase 11 fills this in
	runResolverCheckSummary()
	// RESOLVER_SECTION_END

	// MESH_SECTION_START — Phase 12 fills this in
	runMeshCheckSummary()
	// MESH_SECTION_END

	// COMPUTE_ENGINE_SECTION_START — Phase 13 fills this in
	runComputeCheckSummary()
	// COMPUTE_ENGINE_SECTION_END

	if _, err := os.Stat(workspaceRoot()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("workspace stat %s: %w", workspaceRoot(), err)
	}

	return nil
}

func runGPUCheckSummary() {
	state, err := gpuDetectState()
	if err != nil {
		ui.Fail("GPU", err.Error())
		return
	}

	switch state {
	case gpu.GPUStateNVIDIA:
		ui.Pass("GPU", "NVIDIA detected")
	case gpu.GPUStateAMD:
		ui.Warn("GPU", "AMD detected — ROCm support coming in Gradient Linux v0.3")
	default:
		ui.Info("GPU", "cpu-only")
	}
}

func runGPUDoctorCheck() {
	runGPUCheckSummary()
}

func init() {
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(doctorCmd)
}
