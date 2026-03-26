package cmd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Gradient-Linux/concave/internal/auth"
	"github.com/Gradient-Linux/concave/internal/config"
	"github.com/Gradient-Linux/concave/internal/gpu"
	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/Gradient-Linux/concave/internal/workspace"
	"github.com/spf13/cobra"
)

type mockExitError struct {
	code int
}

func (m mockExitError) Error() string { return fmt.Sprintf("exit %d", m.code) }
func (m mockExitError) ExitCode() int { return m.code }

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func restoreCommandDeps(t *testing.T) {
	t.Helper()

	restoreRole := auth.SetCLIRoleForTesting(auth.RoleAdmin)

	oldExitFunc := exitFunc
	oldEnsureWorkspaceLayout := ensureWorkspaceLayout
	oldWorkspaceExists := workspaceExists
	oldWorkspaceRoot := workspaceRoot
	oldWorkspaceUserRoot := workspaceUserRoot
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
	oldLoadSetupState := loadSetupState
	oldSaveSetupState := saveSetupState
	oldMarkStepComplete := markStepComplete
	oldMarkSetupComplete := markSetupComplete
	oldIsStepComplete := isStepComplete
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
	oldWaitHealthy := waitHealthy
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
	oldSystemCommandAvailable := systemCommandAvailable
	oldSystemDockerCompose := systemDockerCompose
	oldSystemUserInDockerGroup := systemUserInDockerGroup
	oldSystemInternetReachable := systemInternetReachable
	oldSystemRunPrivileged := systemRunPrivileged
	oldSystemCheckConflicts := systemCheckConflicts
	oldSystemRegisterPorts := systemRegisterPorts
	oldSystemDeregisterPorts := systemDeregisterPorts
	oldSystemOpenURL := systemOpenURL
	oldSystemSignalHandler := systemSignalHandler
	oldSystemLock := systemLock
	oldUIConfirm := uiConfirm
	oldCurrentUsername := currentUsername
	oldRunDockerOutput := runDockerOutput
	oldRunDockerInteractive := runDockerInteractive
	oldLabSuite := labSuite
	oldLogsService := logsService
	oldLogsLines := logsLines
	oldLogsFollow := logsFollow
	oldInstallForce := installForce

	t.Cleanup(func() {
		restoreRole()
		exitFunc = oldExitFunc
		ensureWorkspaceLayout = oldEnsureWorkspaceLayout
		workspaceExists = oldWorkspaceExists
		workspaceRoot = oldWorkspaceRoot
		workspaceUserRoot = oldWorkspaceUserRoot
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
		loadSetupState = oldLoadSetupState
		saveSetupState = oldSaveSetupState
		markStepComplete = oldMarkStepComplete
		markSetupComplete = oldMarkSetupComplete
		isStepComplete = oldIsStepComplete
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
		waitHealthy = oldWaitHealthy
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
		systemCommandAvailable = oldSystemCommandAvailable
		systemDockerCompose = oldSystemDockerCompose
		systemUserInDockerGroup = oldSystemUserInDockerGroup
		systemInternetReachable = oldSystemInternetReachable
		systemRunPrivileged = oldSystemRunPrivileged
		systemCheckConflicts = oldSystemCheckConflicts
		systemRegisterPorts = oldSystemRegisterPorts
		systemDeregisterPorts = oldSystemDeregisterPorts
		systemOpenURL = oldSystemOpenURL
		systemSignalHandler = oldSystemSignalHandler
		systemLock = oldSystemLock
		uiConfirm = oldUIConfirm
		currentUsername = oldCurrentUsername
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

	systemSignalHandler = func(parent context.Context, cleanup func()) (context.Context, context.CancelFunc) {
		return context.WithCancel(parent)
	}
	systemLock = func(subcommand string) (func(), error) {
		return func() {}, nil
	}
	waitHealthy = func(ctx context.Context, s suite.Suite, timeout time.Duration, progressFn func([]suite.HealthResult)) error {
		return nil
	}
}

func captureOutput(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	ui.SetOutput(&buf)
	return &buf
}

func setStdin(t *testing.T, input string) {
	t.Helper()
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error = %v", err)
	}
	if _, err := w.WriteString(input); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	_ = w.Close()
	os.Stdin = r
	t.Cleanup(func() {
		_ = r.Close()
		os.Stdin = oldStdin
	})
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

func TestResolveExitCodeFindsWrappedExitErrors(t *testing.T) {
	if code, ok := resolveExitCode(mockExitError{code: 7}); !ok || code != 7 {
		t.Fatalf("resolveExitCode(direct) = %d, %v", code, ok)
	}

	wrapped := fmt.Errorf("outer: %w", mockExitError{code: 19})
	if code, ok := resolveExitCode(wrapped); !ok || code != 19 {
		t.Fatalf("resolveExitCode(wrapped) = %d, %v", code, ok)
	}

	if code, ok := resolveExitCode(errors.New("plain")); ok || code != 0 {
		t.Fatalf("resolveExitCode(plain) = %d, %v", code, ok)
	}
}

func TestInstallInvalidSuite(t *testing.T) {
	restoreCommandDeps(t)

	err := runInstall(installCmd, []string{"invalid"})
	if err == nil || err.Error() != "unknown suite: invalid. Valid suites: boosting, neural, flow, forge" {
		t.Fatalf("runInstall() error = %v", err)
	}
}

func TestEnsureDockerRuntimeInstallsAndStartsDocker(t *testing.T) {
	restoreCommandDeps(t)

	dockerInstalled := false
	dockerRunning := false

	systemCommandAvailable = func(name string) bool {
		return name == "docker" && dockerInstalled
	}
	systemDockerCompose = func() (bool, error) {
		if !dockerInstalled {
			return false, nil
		}
		return true, nil
	}
	systemDockerRunning = func() (bool, error) {
		if !dockerInstalled {
			return false, errors.New("docker info: exec: \"docker\": executable file not found in $PATH")
		}
		return dockerRunning, nil
	}
	systemUserInDockerGroup = func() (bool, error) { return true, nil }
	uiConfirm = func(question string) bool { return true }

	var privileged [][]string
	systemRunPrivileged = func(ctx context.Context, description string, name string, args ...string) error {
		privileged = append(privileged, append([]string{name}, args...))
		if len(args) >= 4 && name == "env" && args[1] == "apt-get" && args[2] == "install" {
			dockerInstalled = true
		}
		if name == "systemctl" {
			dockerRunning = true
		}
		return nil
	}

	if err := ensureDockerRuntime(context.Background(), "install neural"); err != nil {
		t.Fatalf("ensureDockerRuntime() error = %v", err)
	}

	if !dockerInstalled || !dockerRunning {
		t.Fatalf("expected docker to be installed and running, got installed=%v running=%v", dockerInstalled, dockerRunning)
	}
	if len(privileged) < 2 {
		t.Fatalf("expected privileged install/start calls, got %#v", privileged)
	}
}

func TestEnsureDockerRuntimeRequestsReloginAfterDockerGroupAdd(t *testing.T) {
	restoreCommandDeps(t)

	systemCommandAvailable = func(name string) bool { return name == "docker" }
	systemDockerCompose = func() (bool, error) { return true, nil }
	systemDockerRunning = func() (bool, error) { return true, nil }
	systemUserInDockerGroup = func() (bool, error) { return false, nil }
	uiConfirm = func(question string) bool { return true }
	currentUsername = func() string { return "mark" }

	var got []string
	systemRunPrivileged = func(ctx context.Context, description string, name string, args ...string) error {
		got = append([]string{name}, args...)
		return nil
	}

	err := ensureDockerRuntime(context.Background(), "install neural")
	if err == nil {
		t.Fatal("expected relogin error")
	}
	if !strings.Contains(err.Error(), "log out and back in") {
		t.Fatalf("expected relogin guidance, got %v", err)
	}
	if strings.Join(got, " ") != "usermod -aG docker mark" {
		t.Fatalf("unexpected privileged call %#v", got)
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

func TestStartStartsInstalledSuiteAndRegistersPorts(t *testing.T) {
	restoreCommandDeps(t)
	buf := captureOutput(t)

	isInstalled = func(name string) (bool, error) { return true, nil }
	dockerComposePath = func(name string) string { return "/tmp/" + name + ".compose.yml" }

	var composeCalls []string
	dockerComposeUp = func(ctx context.Context, path string, detach bool) error {
		composeCalls = append(composeCalls, path)
		if !detach {
			t.Fatal("expected detached compose up")
		}
		return nil
	}

	registered := ""
	systemRegisterPorts = func(s suite.Suite) error {
		registered = s.Name
		return nil
	}

	if err := runStart(startCmd, []string{"boosting"}); err != nil {
		t.Fatalf("runStart() error = %v", err)
	}
	if len(composeCalls) != 1 || composeCalls[0] != "/tmp/boosting.compose.yml" {
		t.Fatalf("unexpected compose calls %#v", composeCalls)
	}
	if registered != "boosting" {
		t.Fatalf("systemRegisterPorts() suite = %q", registered)
	}
	if !strings.Contains(buf.String(), "Started") {
		t.Fatalf("expected start output, got %q", buf.String())
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

func TestRemoveMissingComposeStillCleansState(t *testing.T) {
	restoreCommandDeps(t)
	buf := captureOutput(t)

	isInstalled = func(name string) (bool, error) { return true, nil }
	uiConfirm = func(question string) bool { return true }
	dockerComposePath = func(name string) string { return filepath.Join(t.TempDir(), name+".compose.yml") }

	var removedState string
	removeSuite = func(name string) error {
		removedState = name
		return nil
	}

	loadManifest = func() (config.VersionManifest, error) {
		return config.VersionManifest{
			"boosting": {
				"gradient-boost-core": {Current: "python:3.12-slim"},
			},
		}, nil
	}

	var saved config.VersionManifest
	saveManifest = func(manifest config.VersionManifest) error {
		saved = manifest
		return nil
	}

	dockerContainerStatus = func(ctx context.Context, name string) (string, error) { return "running", nil }
	runDockerOutput = func(ctx context.Context, args ...string) ([]byte, error) { return nil, nil }
	systemDeregisterPorts = func(s suite.Suite) error { return nil }

	if err := runRemove(removeCmd, []string{"boosting"}); err != nil {
		t.Fatalf("runRemove() error = %v", err)
	}
	if removedState != "boosting" {
		t.Fatalf("removeSuite() name = %q", removedState)
	}
	if _, ok := saved["boosting"]; ok {
		t.Fatalf("expected boosting manifest entry to be removed: %#v", saved)
	}
	if !strings.Contains(buf.String(), "missing — cleaning suite state directly") {
		t.Fatalf("expected missing compose warning, got %q", buf.String())
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

func TestStatusHelpersCoverPortAndFallbackBranches(t *testing.T) {
	restoreCommandDeps(t)

	boosting := suite.Registry["boosting"]
	if got := containerPortSummary(boosting, boosting.Containers[1]); got != "8888" {
		t.Fatalf("containerPortSummary(boosting lab) = %q", got)
	}
	if got := containerPortSummary(boosting, boosting.Containers[2]); got != "5000" {
		t.Fatalf("containerPortSummary(boosting track) = %q", got)
	}

	neural := suite.Registry["neural"]
	if got := containerPortSummary(neural, neural.Containers[1]); got != "8000,8080" {
		t.Fatalf("containerPortSummary(neural infer) = %q", got)
	}

	flow := suite.Registry["flow"]
	cases := map[string]string{
		"gradient-flow-airflow":    "8080",
		"gradient-flow-prometheus": "9090",
		"gradient-flow-grafana":    "3000",
		"gradient-flow-store":      "9001",
		"gradient-flow-serve":      "3100",
	}
	for _, container := range flow.Containers {
		if want, ok := cases[container.Name]; ok {
			if got := containerPortSummary(flow, container); got != want {
				t.Fatalf("containerPortSummary(%s) = %q, want %q", container.Name, got, want)
			}
		}
	}
	if got := containerPortSummary(flow, flow.Containers[0]); got != "5000" {
		t.Fatalf("containerPortSummary(flow mlflow) = %q", got)
	}
	if got := containerPortSummary(suite.Suite{}, suite.Container{Name: "unknown"}); got != "—" {
		t.Fatalf("containerPortSummary(unknown) = %q", got)
	}

	gpuDetectState = func() (gpu.GPUState, error) { return gpu.GPUStateAMD, nil }
	if line, ok := currentGPULine(); !ok || !strings.Contains(line, "AMD detected") {
		t.Fatalf("currentGPULine(AMD) = %q, %v", line, ok)
	}

	gpuDetectState = func() (gpu.GPUState, error) { return gpu.GPUStateNone, nil }
	if line, ok := currentGPULine(); ok || line != "" {
		t.Fatalf("currentGPULine(none) = %q, %v", line, ok)
	}

	workspaceRoot = func() string { return filepath.Join(t.TempDir(), "missing") }
	if line := currentWorkspaceLine(); !strings.Contains(line, "Workspace") {
		t.Fatalf("currentWorkspaceLine() = %q", line)
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
	workspaceUserRoot = func() string { return filepath.Join(home, "gradient") }
	ensureWorkspaceLayout = func() error {
		return os.MkdirAll(filepath.Join(home, "gradient"), 0o755)
	}
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

func TestWorkspaceCommandsPreferUserWorkspaceRoot(t *testing.T) {
	restoreCommandDeps(t)

	home := t.TempDir()
	expected := filepath.Join(home, "gradient")
	t.Setenv("HOME", home)
	t.Setenv("GRADIENT_WORKSPACE_ROOT", "/var/lib/gradient")

	workspaceUserRoot = func() string { return expected }
	ensureWorkspaceLayout = func() error {
		if got := os.Getenv("GRADIENT_WORKSPACE_ROOT"); got != expected {
			t.Fatalf("GRADIENT_WORKSPACE_ROOT = %q, want %q", got, expected)
		}
		return nil
	}
	workspaceStatus = func() ([]workspace.Usage, error) {
		if got := os.Getenv("GRADIENT_WORKSPACE_ROOT"); got != expected {
			t.Fatalf("GRADIENT_WORKSPACE_ROOT = %q, want %q", got, expected)
		}
		return []workspace.Usage{{Name: "data", Bytes: 1}}, nil
	}
	workspaceBackup = func() (string, error) {
		if got := os.Getenv("GRADIENT_WORKSPACE_ROOT"); got != expected {
			t.Fatalf("GRADIENT_WORKSPACE_ROOT = %q, want %q", got, expected)
		}
		return filepath.Join(expected, "backups", "archive.tar.gz"), nil
	}
	workspaceClean = func() error {
		if got := os.Getenv("GRADIENT_WORKSPACE_ROOT"); got != expected {
			t.Fatalf("GRADIENT_WORKSPACE_ROOT = %q, want %q", got, expected)
		}
		return nil
	}
	workspaceRoot = func() string {
		return os.Getenv("GRADIENT_WORKSPACE_ROOT")
	}

	if err := workspaceInitCmd.RunE(workspaceInitCmd, nil); err != nil {
		t.Fatalf("workspaceInitCmd.RunE() error = %v", err)
	}
	if err := workspaceStatusCmd.RunE(workspaceStatusCmd, nil); err != nil {
		t.Fatalf("workspaceStatusCmd.RunE() error = %v", err)
	}
	if err := workspaceBackupCmd.RunE(workspaceBackupCmd, nil); err != nil {
		t.Fatalf("workspaceBackupCmd.RunE() error = %v", err)
	}
	workspacePruneOutputs = true
	defer func() { workspacePruneOutputs = false }()
	if err := workspacePruneCmd.RunE(workspacePruneCmd, nil); err != nil {
		t.Fatalf("workspacePruneCmd.RunE() error = %v", err)
	}

	if got := os.Getenv("GRADIENT_WORKSPACE_ROOT"); got != "/var/lib/gradient" {
		t.Fatalf("GRADIENT_WORKSPACE_ROOT after command = %q, want %q", got, "/var/lib/gradient")
	}
}

func TestConfigureDefaultWorkspaceRoot(t *testing.T) {
	restoreCommandDeps(t)

	workspaceUserRoot = func() string { return "/tmp/user-gradient" }

	_ = os.Unsetenv("GRADIENT_WORKSPACE_ROOT")
	configureDefaultWorkspaceRoot(statusCmd)
	if got := os.Getenv("GRADIENT_WORKSPACE_ROOT"); got != "/tmp/user-gradient" {
		t.Fatalf("GRADIENT_WORKSPACE_ROOT = %q, want %q", got, "/tmp/user-gradient")
	}

	_ = os.Unsetenv("GRADIENT_WORKSPACE_ROOT")
	configureDefaultWorkspaceRoot(serveCmd)
	if got, ok := os.LookupEnv("GRADIENT_WORKSPACE_ROOT"); ok {
		t.Fatalf("GRADIENT_WORKSPACE_ROOT unexpectedly set for serve: %q", got)
	}

	_ = os.Setenv("GRADIENT_WORKSPACE_ROOT", "/explicit/root")
	configureDefaultWorkspaceRoot(statusCmd)
	if got := os.Getenv("GRADIENT_WORKSPACE_ROOT"); got != "/explicit/root" {
		t.Fatalf("explicit GRADIENT_WORKSPACE_ROOT overwritten: %q", got)
	}
}

func TestUpdateRollbackChangelogAndHelpers(t *testing.T) {
	restoreCommandDeps(t)
	buf := captureOutput(t)

	manifest := config.VersionManifest{
		"boosting": {
			"gradient-boost-core":  {Current: "python:3.11-slim", Previous: "python:3.10-slim"},
			"gradient-boost-lab":   {Current: "lab:old", Previous: "lab:older"},
			"gradient-boost-track": {Current: "track:old", Previous: "track:older"},
		},
	}
	isInstalled = func(name string) (bool, error) { return true, nil }
	loadManifest = func() (config.VersionManifest, error) { return manifest, nil }
	saveManifest = func(next config.VersionManifest) error {
		manifest = next
		return nil
	}
	dockerPullWithRollbackSafety = func(ctx context.Context, image string, cb func(string)) error { return nil }
	dockerWriteCompose = func(name string) (string, error) { return "/tmp/" + name + ".compose.yml", nil }
	dockerComposeUp = func(ctx context.Context, path string, detach bool) error { return nil }
	dockerComposeDown = func(ctx context.Context, path string) error { return nil }
	dockerComposePath = func(name string) string { return "/tmp/" + name + ".compose.yml" }
	dockerContainerStatus = func(ctx context.Context, name string) (string, error) { return "running", nil }

	if err := runUpdate(updateCmd, []string{"boosting"}); err != nil {
		t.Fatalf("runUpdate() error = %v", err)
	}
	if err := runRollback(rollbackCmd, []string{"boosting"}); err != nil {
		t.Fatalf("runRollback() error = %v", err)
	}
	if err := runChangelog(changelogCmd, []string{"boosting"}); err != nil {
		t.Fatalf("runChangelog() error = %v", err)
	}

	if manifest["boosting"]["gradient-boost-core"].Current == "" {
		t.Fatalf("expected manifest to stay populated, got %#v", manifest)
	}
	for _, token := range []string{"Update", "Rollback", "suite changelog"} {
		if !strings.Contains(buf.String(), token) {
			t.Fatalf("expected %q in output %q", token, buf.String())
		}
	}
}

func TestRemoveStopRestartAndShellCommands(t *testing.T) {
	restoreCommandDeps(t)

	isInstalled = func(name string) (bool, error) { return true, nil }
	uiConfirm = func(question string) bool { return true }
	manifest := config.VersionManifest{
		"boosting": {
			"gradient-boost-core":  {Current: "python:3.12-slim"},
			"gradient-boost-lab":   {Current: "quay.io/jupyter/base-notebook:python-3.11.6"},
			"gradient-boost-track": {Current: "ghcr.io/mlflow/mlflow:v2.14.1"},
		},
	}
	loadManifest = func() (config.VersionManifest, error) { return manifest, nil }
	saveManifest = func(next config.VersionManifest) error {
		manifest = next
		return nil
	}
	loadState = func() (config.State, error) { return config.State{Installed: []string{"boosting", "flow"}}, nil }
	dockerComposePath = func(name string) string { return "/tmp/" + name + ".compose.yml" }
	removeSuite = func(name string) error { return nil }
	systemDeregisterPorts = func(s suite.Suite) error { return nil }
	systemRegisterPorts = func(s suite.Suite) error { return nil }
	dockerContainerStatus = func(ctx context.Context, name string) (string, error) { return "running", nil }

	var interactiveCalls []string
	runDockerInteractive = func(ctx context.Context, args ...string) error {
		call := strings.Join(args, " ")
		interactiveCalls = append(interactiveCalls, call)
		if strings.Contains(call, " bash") {
			return fmt.Errorf("bash missing")
		}
		return nil
	}
	runDockerOutput = func(ctx context.Context, args ...string) ([]byte, error) { return nil, nil }
	dockerComposeDown = func(ctx context.Context, path string) error { return nil }
	dockerComposeUp = func(ctx context.Context, path string, detach bool) error { return nil }

	if err := runRemove(removeCmd, []string{"boosting"}); err != nil {
		t.Fatalf("runRemove() error = %v", err)
	}
	if err := runStop(stopCmd, nil); err != nil {
		t.Fatalf("runStop() error = %v", err)
	}
	if err := runRestart(restartCmd, []string{"boosting"}); err != nil {
		t.Fatalf("runRestart() error = %v", err)
	}
	if err := runShell(shellCmd, []string{"boosting"}); err != nil {
		t.Fatalf("runShell() error = %v", err)
	}
	if len(interactiveCalls) == 0 {
		t.Fatal("expected interactive docker calls")
	}
	if _, ok := manifest["boosting"]; ok {
		t.Fatalf("expected boosting manifest entry to be removed, got %#v", manifest)
	}
}

func TestHelperFunctionsAndFallbackPaths(t *testing.T) {
	restoreCommandDeps(t)

	loadState = func() (config.State, error) { return config.State{Installed: []string{"flow", "boosting"}}, nil }
	names, err := installedSuiteTargets(nil, false)
	if err != nil {
		t.Fatalf("installedSuiteTargets() error = %v", err)
	}
	if strings.Join(names, ",") != "boosting,flow" {
		t.Fatalf("unexpected ordered suites %#v", names)
	}

	loadManifest = func() (config.VersionManifest, error) {
		return config.VersionManifest{
			"forge": {
				"gradient-boost-core":  {Current: "python:3.12-slim"},
				"gradient-boost-lab":   {Current: "quay.io/jupyter/base-notebook:python-3.11.6"},
				"gradient-flow-mlflow": {Current: "ghcr.io/mlflow/mlflow:v2.14.1"},
			},
		}, nil
	}
	rawComposeWritten := false
	dockerWriteRawCompose = func(name string, data []byte) (string, error) {
		rawComposeWritten = strings.Contains(string(data), "gradient-boost-core")
		return "/tmp/" + name + ".compose.yml", nil
	}
	forgeSuite, err := currentSuiteDefinition("forge")
	if err != nil {
		t.Fatalf("currentSuiteDefinition() error = %v", err)
	}
	if len(containerNames(forgeSuite)) == 0 {
		t.Fatalf("expected forge containers, got %#v", forgeSuite)
	}
	if _, err := writeComposeForCurrentState("forge"); err != nil {
		t.Fatalf("writeComposeForCurrentState() error = %v", err)
	}
	if !rawComposeWritten {
		t.Fatal("expected forge compose to be rendered through raw writer")
	}

	rawURL, err := extractLabURL("http://0.0.0.0:8888/?token=fallback :: /notebooks")
	if err != nil {
		t.Fatalf("extractLabURL() fallback error = %v", err)
	}
	if rawURL != "http://127.0.0.1:8888/lab?token=fallback" {
		t.Fatalf("unexpected fallback URL %q", rawURL)
	}
}

func TestGPUAndWorkspaceStatusHelpers(t *testing.T) {
	restoreCommandDeps(t)

	tmp := t.TempDir()
	script := filepath.Join(tmp, "nvidia-smi")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho 'NVIDIA RTX 4090, 24576 MiB'\n"), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("PATH", tmp+string(os.PathListSeparator)+os.Getenv("PATH"))
	gpuDetectState = func() (gpu.GPUState, error) { return gpu.GPUStateNVIDIA, nil }
	workspaceRoot = func() string { return tmp }

	line, ok := currentGPULine()
	if !ok || !strings.Contains(line, "NVIDIA RTX 4090") {
		t.Fatalf("currentGPULine() = %q, %v", line, ok)
	}

	workspaceLine := currentWorkspaceLine()
	if !strings.Contains(workspaceLine, "Workspace") {
		t.Fatalf("unexpected workspace line %q", workspaceLine)
	}
}

func TestDriverWizardSetupAndSelfUpdate(t *testing.T) {
	restoreCommandDeps(t)
	buf := captureOutput(t)

	gpuDetectState = func() (gpu.GPUState, error) { return gpu.GPUStateNone, nil }
	if err := driverWizardCmd.RunE(driverWizardCmd, nil); err != nil {
		t.Fatalf("driverWizardCmd cpu-only error = %v", err)
	}

	gpuDetectState = func() (gpu.GPUState, error) { return gpu.GPUStateAMD, nil }
	gpuDetectAMDState = func() gpu.GPUState { return gpu.GPUStateAMD }
	if err := driverWizardCmd.RunE(driverWizardCmd, nil); err != nil {
		t.Fatalf("driverWizardCmd AMD error = %v", err)
	}

	gpuDetectState = func() (gpu.GPUState, error) { return gpu.GPUStateNVIDIA, nil }
	gpuSecureBootEnabled = func() (bool, error) { return true, nil }
	gpuRecommendedDriverBranch = func() (string, error) { return "570", nil }
	gpuToolkitConfigured = func() (bool, error) { return true, nil }
	gpuVerifyPassthrough = func() error { return nil }
	setStdin(t, "n\n")
	if err := driverWizardCmd.RunE(driverWizardCmd, nil); err != nil {
		t.Fatalf("driverWizardCmd secure boot error = %v", err)
	}

	ensureWorkspaceLayout = func() error { return nil }
	home := t.TempDir()
	t.Setenv("HOME", home)
	workspaceRoot = func() string { return home }
	systemCommandAvailable = func(name string) bool { return name == "docker" }
	systemDockerCompose = func() (bool, error) { return true, nil }
	systemDockerRunning = func() (bool, error) { return true, nil }
	systemUserInDockerGroup = func() (bool, error) { return true, nil }
	gpuDetectState = func() (gpu.GPUState, error) { return gpu.GPUStateNone, nil }
	installed := []string{}
	isInstalled = func(name string) (bool, error) {
		for _, item := range installed {
			if item == name {
				return true, nil
			}
		}
		return false, nil
	}
	installSuite = func(ctx context.Context, name string, opts suite.InstallOptions) error {
		installed = append(installed, name)
		return nil
	}
	dockerComposePath = func(name string) string { return "/tmp/" + name + ".compose.yml" }
	dockerComposeUp = func(ctx context.Context, path string, detach bool) error { return nil }
	systemRegisterPorts = func(s suite.Suite) error { return nil }
	doctorCalled := false
	doctorCmd.RunE = func(cmd *cobra.Command, args []string) error {
		doctorCalled = true
		return nil
	}
	setStdin(t, "\n")
	if err := setupCmd.RunE(setupCmd, nil); err != nil {
		t.Fatalf("setupCmd.RunE() error = %v", err)
	}
	if len(installed) != 1 || installed[0] != "boosting" || !doctorCalled {
		t.Fatalf("unexpected setup behavior installed=%#v doctor=%v", installed, doctorCalled)
	}

	binary := []byte("concave-binary")
	sum := sha256.Sum256(binary)
	const manifestURL = "https://updates.example.test/manifest"
	const binaryURL = "https://updates.example.test/concave"
	selfUpdateClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.String() {
			case manifestURL:
				payload, err := json.Marshal(updateManifest{
					Version: "v0.1.0",
					URL:     binaryURL,
					SHA256:  hex.EncodeToString(sum[:]),
				})
				if err != nil {
					return nil, err
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(string(payload))),
					Header:     make(http.Header),
				}, nil
			case binaryURL:
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(string(binary))),
					Header:     make(http.Header),
				}, nil
			default:
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("not found")),
					Header:     make(http.Header),
				}, nil
			}
		}),
	}
	selfUpdateManifestURL = manifestURL
	selfUpdateTargetPath = filepath.Join(t.TempDir(), "concave")
	if err := selfUpdateCmd.RunE(selfUpdateCmd, nil); err != nil {
		t.Fatalf("selfUpdateCmd.RunE() error = %v", err)
	}
	data, err := os.ReadFile(selfUpdateTargetPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != string(binary) {
		t.Fatalf("unexpected updated binary %q", string(data))
	}
	if !strings.Contains(buf.String(), "Updated") {
		t.Fatalf("expected update output, got %q", buf.String())
	}
}

func filepathJoin(elem ...string) string {
	return filepath.Join(elem...)
}
