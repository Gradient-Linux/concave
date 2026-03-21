package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Gradient-Linux/concave/internal/config"
	"github.com/Gradient-Linux/concave/internal/gpu"
	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/Gradient-Linux/concave/internal/ui"
)

type mockExitError struct {
	code int
}

func (m mockExitError) Error() string { return fmt.Sprintf("exit %d", m.code) }
func (m mockExitError) ExitCode() int { return m.code }

func restoreCommandDeps(t *testing.T) {
	t.Helper()

	oldExitFunc := exitFunc
	oldEnsureWorkspaceLayout := ensureWorkspaceLayout
	oldWorkspaceExists := workspaceExists
	oldWorkspaceRoot := workspaceRoot
	oldWorkspaceStatus := workspaceStatus
	oldWorkspaceBackup := workspaceBackup
	oldWorkspaceClean := workspaceClean
	oldLoadState := loadState
	oldSaveState := saveState
	oldAddSuite := addSuite
	oldRemoveSuite := removeSuite
	oldIsInstalled := isInstalled
	oldLoadManifest := loadManifest
	oldSaveManifest := saveManifest
	oldRecordInstall := recordInstall
	oldRecordUpdate := recordUpdate
	oldSwapForRollback := swapForRollback
	oldGetSuite := getSuite
	oldAllSuites := allSuites
	oldSuiteNames := suiteNames
	oldPrimaryContainer := primaryContainer
	oldJupyterContainer := jupyterContainer
	oldPickForgeComponents := pickForgeComponents
	oldBuildForgeCompose := buildForgeCompose
	oldForgeSelectionFromNames := forgeSelectionFromNames
	oldInstallSuite := installSuite
	oldDockerPullWithRollbackSafety := dockerPullWithRollbackSafety
	oldDockerWriteCompose := dockerWriteCompose
	oldDockerWriteRawCompose := dockerWriteRawCompose
	oldDockerComposePath := dockerComposePath
	oldDockerComposeUp := dockerComposeUp
	oldDockerComposeDown := dockerComposeDown
	oldDockerContainerStatus := dockerContainerStatus
	oldDockerContainerLogs := dockerContainerLogs
	oldDockerExecCommand := dockerExecCommand
	oldDockerRevertToPrevious := dockerRevertToPrevious
	oldGPUDetectState := gpuDetectState
	oldGPUDetectAMDState := gpuDetectAMDState
	oldGPURecommendedDriverBranch := gpuRecommendedDriverBranch
	oldGPUToolkitConfigured := gpuToolkitConfigured
	oldGPUVerifyPassthrough := gpuVerifyPassthrough
	oldGPUSecureBootEnabled := gpuSecureBootEnabled
	oldSystemDockerRunning := systemDockerRunning
	oldSystemUserInDockerGroup := systemUserInDockerGroup
	oldSystemInternetReachable := systemInternetReachable
	oldSystemCheckConflicts := systemCheckConflicts
	oldSystemRegisterPorts := systemRegisterPorts
	oldSystemDeregisterPorts := systemDeregisterPorts
	oldSystemOpenURL := systemOpenURL
	oldUIConfirm := uiConfirm
	oldRunDockerOutput := runDockerOutput
	oldRunDockerInteractive := runDockerInteractive
	oldLabSuite := labSuite
	oldLogsService := logsService
	oldLogsLines := logsLines
	oldLogsFollow := logsFollow
	oldInstallForce := installForce

	t.Cleanup(func() {
		exitFunc = oldExitFunc
		ensureWorkspaceLayout = oldEnsureWorkspaceLayout
		workspaceExists = oldWorkspaceExists
		workspaceRoot = oldWorkspaceRoot
		workspaceStatus = oldWorkspaceStatus
		workspaceBackup = oldWorkspaceBackup
		workspaceClean = oldWorkspaceClean
		loadState = oldLoadState
		saveState = oldSaveState
		addSuite = oldAddSuite
		removeSuite = oldRemoveSuite
		isInstalled = oldIsInstalled
		loadManifest = oldLoadManifest
		saveManifest = oldSaveManifest
		recordInstall = oldRecordInstall
		recordUpdate = oldRecordUpdate
		swapForRollback = oldSwapForRollback
		getSuite = oldGetSuite
		allSuites = oldAllSuites
		suiteNames = oldSuiteNames
		primaryContainer = oldPrimaryContainer
		jupyterContainer = oldJupyterContainer
		pickForgeComponents = oldPickForgeComponents
		buildForgeCompose = oldBuildForgeCompose
		forgeSelectionFromNames = oldForgeSelectionFromNames
		installSuite = oldInstallSuite
		dockerPullWithRollbackSafety = oldDockerPullWithRollbackSafety
		dockerWriteCompose = oldDockerWriteCompose
		dockerWriteRawCompose = oldDockerWriteRawCompose
		dockerComposePath = oldDockerComposePath
		dockerComposeUp = oldDockerComposeUp
		dockerComposeDown = oldDockerComposeDown
		dockerContainerStatus = oldDockerContainerStatus
		dockerContainerLogs = oldDockerContainerLogs
		dockerExecCommand = oldDockerExecCommand
		dockerRevertToPrevious = oldDockerRevertToPrevious
		gpuDetectState = oldGPUDetectState
		gpuDetectAMDState = oldGPUDetectAMDState
		gpuRecommendedDriverBranch = oldGPURecommendedDriverBranch
		gpuToolkitConfigured = oldGPUToolkitConfigured
		gpuVerifyPassthrough = oldGPUVerifyPassthrough
		gpuSecureBootEnabled = oldGPUSecureBootEnabled
		systemDockerRunning = oldSystemDockerRunning
		systemUserInDockerGroup = oldSystemUserInDockerGroup
		systemInternetReachable = oldSystemInternetReachable
		systemCheckConflicts = oldSystemCheckConflicts
		systemRegisterPorts = oldSystemRegisterPorts
		systemDeregisterPorts = oldSystemDeregisterPorts
		systemOpenURL = oldSystemOpenURL
		uiConfirm = oldUIConfirm
		runDockerOutput = oldRunDockerOutput
		runDockerInteractive = oldRunDockerInteractive
		labSuite = oldLabSuite
		logsService = oldLogsService
		logsLines = oldLogsLines
		logsFollow = oldLogsFollow
		installForce = oldInstallForce
		rootCmd.SetArgs(nil)
		ui.ResetOutput()
	})
}

func captureOutput(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	ui.SetOutput(&buf)
	return &buf
}

func TestExecutePreservesExitCode(t *testing.T) {
	restoreCommandDeps(t)
	buf := captureOutput(t)

	getSuite = func(name string) (suite.Suite, error) { return suite.Registry["boosting"], nil }
	isInstalled = func(name string) (bool, error) { return true, nil }
	dockerContainerStatus = func(ctx context.Context, name string) (string, error) { return "running", nil }
	runDockerInteractive = func(ctx context.Context, args ...string) error { return mockExitError{code: 42} }

	exitCode := 0
	exitFunc = func(code int) { exitCode = code }

	rootCmd.SetArgs([]string{"exec", "boosting", "--", "python", "-c", "raise SystemExit(42)"})
	Execute()

	if exitCode != 42 {
		t.Fatalf("Execute() exit code = %d, want 42", exitCode)
	}
	if !strings.Contains(buf.String(), "Error") {
		t.Fatalf("expected error output, got %q", buf.String())
	}
}

func TestInstallInvalidSuite(t *testing.T) {
	restoreCommandDeps(t)

	gpuDetectState = func() (gpu.GPUState, error) { return gpu.GPUStateNone, nil }
	installSuite = func(ctx context.Context, name string, opts suite.InstallOptions) error {
		_, err := suite.Get(name)
		return err
	}

	err := runInstall(installCmd, []string{"invalid"})
	if err == nil || err.Error() != "unknown suite: invalid. Valid suites: boosting, neural, flow, forge" {
		t.Fatalf("runInstall() error = %v", err)
	}
}

func TestStartWithNoInstalledSuites(t *testing.T) {
	restoreCommandDeps(t)
	buf := captureOutput(t)

	loadState = func() (config.State, error) { return config.State{Installed: []string{}}, nil }
	if err := runStart(startCmd, nil); err != nil {
		t.Fatalf("runStart() error = %v", err)
	}
	if !strings.Contains(buf.String(), "No suites installed. Run: concave install [suite]") {
		t.Fatalf("unexpected output %q", buf.String())
	}
}

func TestRemoveNegativeConfirmationNoOp(t *testing.T) {
	restoreCommandDeps(t)

	isInstalled = func(name string) (bool, error) { return true, nil }
	uiConfirm = func(question string) bool { return false }
	removed := false
	removeSuite = func(name string) error {
		removed = true
		return nil
	}

	if err := runRemove(removeCmd, []string{"boosting"}); err != nil {
		t.Fatalf("runRemove() error = %v", err)
	}
	if removed {
		t.Fatal("expected removal to be skipped")
	}
}

func TestLabPrefersBoostingWhenBothInstalled(t *testing.T) {
	restoreCommandDeps(t)
	buf := captureOutput(t)

	loadState = func() (config.State, error) { return config.State{Installed: []string{"neural", "boosting"}}, nil }
	isInstalled = func(name string) (bool, error) { return true, nil }
	dockerContainerStatus = func(ctx context.Context, name string) (string, error) { return "running", nil }
	runDockerOutput = func(ctx context.Context, args ...string) ([]byte, error) {
		if strings.Join(args, " ") != "exec gradient-boost-lab jupyter server list --json" {
			t.Fatalf("unexpected docker args %q", strings.Join(args, " "))
		}
		return []byte("{\"url\":\"http://localhost:8888/\",\"token\":\"abc123\"}\n"), nil
	}
	opened := ""
	systemOpenURL = func(url string) error { opened = url; return nil }

	if err := runLab(labCmd, nil); err != nil {
		t.Fatalf("runLab() error = %v", err)
	}
	if opened != "http://localhost:8888/lab?token=abc123" {
		t.Fatalf("unexpected opened URL %q", opened)
	}
	if !strings.Contains(buf.String(), "opening at http://localhost:8888/lab?token=abc123") {
		t.Fatalf("expected printed URL, got %q", buf.String())
	}
}

func TestLogsBuildsComposeInvocation(t *testing.T) {
	restoreCommandDeps(t)

	isInstalled = func(name string) (bool, error) { return true, nil }
	dockerComposePath = func(name string) string { return "/tmp/" + name + ".compose.yml" }
	logsService = "gradient-boost-lab"
	logsLines = 25
	logsFollow = false

	var got []string
	runDockerInteractive = func(ctx context.Context, args ...string) error {
		got = append([]string(nil), args...)
		return nil
	}

	if err := runLogs(logsCmd, []string{"boosting"}); err != nil {
		t.Fatalf("runLogs() error = %v", err)
	}

	want := []string{"compose", "-f", "/tmp/boosting.compose.yml", "logs", "--tail", "25", "gradient-boost-lab"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("runLogs() args = %q, want %q", strings.Join(got, " "), strings.Join(want, " "))
	}
}

func TestStatusAndListRenderCurrentState(t *testing.T) {
	restoreCommandDeps(t)
	buf := captureOutput(t)

	loadState = func() (config.State, error) { return config.State{Installed: []string{"boosting"}}, nil }
	loadManifest = func() (config.VersionManifest, error) {
		return config.VersionManifest{
			"boosting": {
				"gradient-boost-core": {Current: "python:3.12-slim", Previous: ""},
			},
		}, nil
	}
	dockerContainerStatus = func(ctx context.Context, name string) (string, error) { return "running", nil }
	gpuDetectState = func() (gpu.GPUState, error) { return gpu.GPUStateNone, nil }
	workspaceRoot = func() string { return t.TempDir() }

	if err := runList(listCmd, nil); err != nil {
		t.Fatalf("runList() error = %v", err)
	}
	if err := runStatus(statusCmd, nil); err != nil {
		t.Fatalf("runStatus() error = %v", err)
	}

	for _, token := range []string{"Installed Suites", "boosting", "python:3.12-slim", "Gradient Linux — Suite Status", "gradient-boost-core", "running"} {
		if !strings.Contains(buf.String(), token) {
			t.Fatalf("expected %q in output %q", token, buf.String())
		}
	}
}

func TestDoctorAndWorkspaceCommandsStillWork(t *testing.T) {
	restoreCommandDeps(t)

	buf := captureOutput(t)
	systemDockerRunning = func() (bool, error) { return true, nil }
	systemUserInDockerGroup = func() (bool, error) { return true, nil }
	systemInternetReachable = func() (bool, error) { return true, nil }
	workspaceExists = func() bool { return true }
	workspaceRoot = func() string { return t.TempDir() }
	gpuDetectState = func() (gpu.GPUState, error) { return gpu.GPUStateNone, nil }

	if err := doctorCmd.RunE(doctorCmd, nil); err != nil {
		t.Fatalf("doctorCmd.RunE() error = %v", err)
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := workspaceInitCmd.RunE(workspaceInitCmd, nil); err != nil {
		t.Fatalf("workspaceInitCmd.RunE() error = %v", err)
	}
	if _, err := os.Stat(filepathJoin(home, "gradient")); err != nil {
		t.Fatalf("workspace init did not create root: %v", err)
	}
	if !strings.Contains(buf.String(), "Docker") {
		t.Fatalf("expected doctor output, got %q", buf.String())
	}
}

func filepathJoin(elem ...string) string {
	return strings.Join(elem, string(os.PathSeparator))
}
