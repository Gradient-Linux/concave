package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/Gradient-Linux/concave/internal/gpu"
	"github.com/Gradient-Linux/concave/internal/ui"
)

func TestRunGPUDoctorCheckStates(t *testing.T) {
	restoreCommandDeps(t)

	tests := []struct {
		name  string
		state gpu.GPUState
		err   error
		want  string
	}{
		{name: "nvidia", state: gpu.GPUStateNVIDIA, want: "NVIDIA detected"},
		{name: "amd", state: gpu.GPUStateAMD, want: "ROCm support coming in Gradient Linux v0.3"},
		{name: "cpu", state: gpu.GPUStateNone, want: "cpu-only"},
		{name: "error", err: errors.New("gpu probe failed"), want: "gpu probe failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ui.SetOutput(&buf)
			defer ui.ResetOutput()

			gpuDetectState = func() (gpu.GPUState, error) {
				if tt.err != nil {
					return gpu.GPUStateNone, tt.err
				}
				return tt.state, nil
			}

			runGPUDoctorCheck()

			if !strings.Contains(buf.String(), tt.want) {
				t.Fatalf("output = %q, want substring %q", buf.String(), tt.want)
			}
		})
	}
}

func TestConfirmSecureBootFlow(t *testing.T) {
	restoreCommandDeps(t)

	tests := []struct {
		name        string
		secureBoot  bool
		checkErr    error
		confirm     bool
		wantProceed bool
		wantOutput  string
	}{
		{name: "disabled", secureBoot: false, wantProceed: true, wantOutput: "disabled"},
		{name: "probe error", checkErr: errors.New("mokutil missing"), wantProceed: true, wantOutput: "mokutil missing"},
		{name: "enabled continue", secureBoot: true, confirm: true, wantProceed: true, wantOutput: "Option A"},
		{name: "enabled exit", secureBoot: true, confirm: false, wantProceed: false, wantOutput: "exit selected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ui.SetOutput(&buf)
			defer ui.ResetOutput()

			gpuSecureBootEnabled = func() (bool, error) {
				if tt.checkErr != nil {
					return false, tt.checkErr
				}
				return tt.secureBoot, nil
			}
			uiConfirm = func(question string) bool { return tt.confirm }

			proceed, err := confirmSecureBootFlow()
			if err != nil {
				t.Fatalf("confirmSecureBootFlow() error = %v", err)
			}
			if proceed != tt.wantProceed {
				t.Fatalf("confirmSecureBootFlow() = %v, want %v", proceed, tt.wantProceed)
			}
			if !strings.Contains(buf.String(), tt.wantOutput) {
				t.Fatalf("output = %q, want substring %q", buf.String(), tt.wantOutput)
			}
		})
	}
}

func TestRunDriverWizardNVIDIAPaths(t *testing.T) {
	restoreCommandDeps(t)

	tests := []struct {
		name        string
		state       gpu.GPUState
		confirm     bool
		passthrough error
		wantErr     string
		wantOutput  string
	}{
		{name: "cpu-only", state: gpu.GPUStateNone, wantOutput: "no driver changes required"},
		{name: "amd", state: gpu.GPUStateAMD, wantOutput: "ROCm support coming in Gradient Linux v0.3"},
		{name: "nvidia success", state: gpu.GPUStateNVIDIA, confirm: true, wantOutput: "docker passthrough verified"},
		{name: "nvidia passthrough failure", state: gpu.GPUStateNVIDIA, confirm: true, passthrough: errors.New("docker failed"), wantErr: "gpu passthrough verification: docker failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ui.SetOutput(&buf)
			defer ui.ResetOutput()

			gpuDetectState = func() (gpu.GPUState, error) { return tt.state, nil }
			gpuDetectAMDState = func() gpu.GPUState {
				ui.Warn("AMD GPU", "detected — ROCm support coming in Gradient Linux v0.3")
				return gpu.GPUStateAMD
			}
			gpuSecureBootEnabled = func() (bool, error) { return false, nil }
			uiConfirm = func(question string) bool { return tt.confirm }
			gpuRecommendedDriverBranch = func() (string, error) { return "570", nil }
			gpuToolkitConfigured = func() (bool, error) { return true, nil }
			gpuVerifyPassthrough = func() error { return tt.passthrough }

			err := runDriverWizard(driverWizardCmd, nil)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("runDriverWizard() error = %v", err)
			}
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("runDriverWizard() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if !strings.Contains(buf.String(), tt.wantOutput) {
				t.Fatalf("output = %q, want substring %q", buf.String(), tt.wantOutput)
			}
		})
	}
}

func TestRunDriverWizardToolkitWarningStillContinues(t *testing.T) {
	restoreCommandDeps(t)

	var buf bytes.Buffer
	ui.SetOutput(&buf)
	defer ui.ResetOutput()

	gpuDetectState = func() (gpu.GPUState, error) { return gpu.GPUStateNVIDIA, nil }
	gpuSecureBootEnabled = func() (bool, error) { return true, nil }
	uiConfirm = func(question string) bool { return true }
	gpuRecommendedDriverBranch = func() (string, error) { return "560", nil }
	gpuToolkitConfigured = func() (bool, error) { return false, errors.New("toolkit missing") }
	gpuVerifyPassthrough = func() error { return nil }

	if err := runDriverWizard(driverWizardCmd, nil); err != nil {
		t.Fatalf("runDriverWizard() error = %v", err)
	}

	for _, token := range []string{"Option A", "Option B", "560", "toolkit missing", "docker passthrough verified"} {
		if !strings.Contains(buf.String(), token) {
			t.Fatalf("output = %q, missing %q", buf.String(), token)
		}
	}
}
