package cmd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gradient-Linux/concave/internal/config"
	"github.com/Gradient-Linux/concave/internal/gpu"
	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/Gradient-Linux/concave/internal/system"
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/Gradient-Linux/concave/internal/workspace"
	"github.com/spf13/cobra"
)

func restoreCommandDeps(t *testing.T) {
	t.Helper()

	oldExitFunc := exitFunc
	oldEnsureWorkspaceLayout := ensureWorkspaceLayout
	oldWorkspaceExists := workspaceExists
	oldWorkspaceRoot := workspaceRoot
	oldWorkspaceStatus := workspaceStatus
	oldWorkspaceBackup := workspaceBackup
	oldWorkspaceClean := workspaceClean
	oldWorkspaceComposePath := workspaceComposePath
	oldLoadState := loadState
	oldAddInstalledSuite := addInstalledSuite
	oldRemoveInstalledSuite := removeInstalledSuite
	oldLoadVersions := loadVersions
	oldSaveVersions := saveVersions
	oldGetImageVersion := getImageVersion
	oldSetImageVersion := setImageVersion
	oldRemoveSuiteVersions := removeSuiteVersions
	oldSwapPreviousVersions := swapPreviousVersions
	oldGetSuite := getSuite
	oldBuildInstallPlan := buildInstallPlan
	oldSelectForgeComponents := selectForgeComponents
	oldBuildForgeCompose := buildForgeCompose
	oldSuiteNames := suiteNames
	oldPrimaryContainer := primaryContainer
	oldJupyterContainer := jupyterContainer
	oldSuitePorts := suitePorts
	oldDockerPullWithProgress := dockerPullWithProgress
	oldDockerWriteSuiteCompose := dockerWriteSuiteCompose
	oldDockerWriteRawCompose := dockerWriteRawCompose
	oldDockerComposeUp := dockerComposeUp
	oldDockerComposeDown := dockerComposeDown
	oldDockerContainerStatus := dockerContainerStatus
	oldDockerExecCommand := dockerExecCommand
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
	oldRunDockerOutput := runDockerOutput
	oldRunDockerInteractive := runDockerInteractive
	oldManifestURL := selfUpdateManifestURL
	oldSelfUpdateClient := selfUpdateClient
	oldSelfUpdateTargetPath := selfUpdateTargetPath
	oldLabSuite := labSuite
	oldLogsService := logsService
	oldWorkspaceCleanOutputs := workspaceCleanOutputs
	oldInstallRunE := installCmd.RunE
	oldDoctorRunE := doctorCmd.RunE

	t.Cleanup(func() {
		exitFunc = oldExitFunc
		ensureWorkspaceLayout = oldEnsureWorkspaceLayout
		workspaceExists = oldWorkspaceExists
		workspaceRoot = oldWorkspaceRoot
		workspaceStatus = oldWorkspaceStatus
		workspaceBackup = oldWorkspaceBackup
		workspaceClean = oldWorkspaceClean
		workspaceComposePath = oldWorkspaceComposePath
		loadState = oldLoadState
		addInstalledSuite = oldAddInstalledSuite
		removeInstalledSuite = oldRemoveInstalledSuite
		loadVersions = oldLoadVersions
		saveVersions = oldSaveVersions
		getImageVersion = oldGetImageVersion
		setImageVersion = oldSetImageVersion
		removeSuiteVersions = oldRemoveSuiteVersions
		swapPreviousVersions = oldSwapPreviousVersions
		getSuite = oldGetSuite
		buildInstallPlan = oldBuildInstallPlan
		selectForgeComponents = oldSelectForgeComponents
		buildForgeCompose = oldBuildForgeCompose
		suiteNames = oldSuiteNames
		primaryContainer = oldPrimaryContainer
		jupyterContainer = oldJupyterContainer
		suitePorts = oldSuitePorts
		dockerPullWithProgress = oldDockerPullWithProgress
		dockerWriteSuiteCompose = oldDockerWriteSuiteCompose
		dockerWriteRawCompose = oldDockerWriteRawCompose
		dockerComposeUp = oldDockerComposeUp
		dockerComposeDown = oldDockerComposeDown
		dockerContainerStatus = oldDockerContainerStatus
		dockerExecCommand = oldDockerExecCommand
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
		runDockerOutput = oldRunDockerOutput
		runDockerInteractive = oldRunDockerInteractive
		selfUpdateManifestURL = oldManifestURL
		selfUpdateClient = oldSelfUpdateClient
		selfUpdateTargetPath = oldSelfUpdateTargetPath
		labSuite = oldLabSuite
		logsService = oldLogsService
		workspaceCleanOutputs = oldWorkspaceCleanOutputs
		installCmd.RunE = oldInstallRunE
		doctorCmd.RunE = oldDoctorRunE
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

func TestExtractLabURL(t *testing.T) {
	raw := "Currently running servers:\nhttp://localhost:8888/?token=abcdef :: /notebooks\n"
	got, err := extractLabURL(raw)
	if err != nil {
		t.Fatalf("extractLabURL() error = %v", err)
	}
	want := "http://127.0.0.1:8888/lab?token=abcdef"
	if got != want {
		t.Fatalf("extractLabURL() = %q, want %q", got, want)
	}
}

func TestTargetSuitesUsesInstalledState(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := workspace.EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout() error = %v", err)
	}
	if err := config.AddInstalled("boosting"); err != nil {
		t.Fatalf("AddInstalled() error = %v", err)
	}
	if err := config.AddInstalled("flow"); err != nil {
		t.Fatalf("AddInstalled() error = %v", err)
	}

	names, err := targetSuites(nil)
	if err != nil {
		t.Fatalf("targetSuites() error = %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 suites, got %d", len(names))
	}
}

func TestExecuteDoctorAndWorkspaceCommands(t *testing.T) {
	restoreCommandDeps(t)
	buf := captureOutput(t)

	exitCode := 0
	exitFunc = func(code int) { exitCode = code }
	rootCmd.SetArgs([]string{"definitely-invalid"})
	Execute()
	if exitCode != 1 {
		t.Fatalf("Execute() exit code = %d", exitCode)
	}
	if !strings.Contains(buf.String(), "Error") {
		t.Fatalf("expected error output, got %q", buf.String())
	}

	buf.Reset()
	tempRoot := t.TempDir()
	systemDockerRunning = func() (bool, error) { return true, nil }
	systemUserInDockerGroup = func() (bool, error) { return false, nil }
	systemInternetReachable = func() (bool, error) { return true, nil }
	workspaceExists = func() bool { return true }
	workspaceRoot = func() string { return tempRoot }
	gpuDetectState = func() (gpu.GPUState, error) { return gpu.GPUStateAMD, nil }
	if err := doctorCmd.RunE(doctorCmd, nil); err != nil {
		t.Fatalf("doctorCmd.RunE() error = %v", err)
	}
	for _, token := range []string{"Docker", "Docker group", "Internet", "Workspace", "AMD"} {
		if !strings.Contains(buf.String(), token) {
			t.Fatalf("doctor output missing %q in %q", token, buf.String())
		}
	}

	buf.Reset()
	t.Setenv("HOME", t.TempDir())
	if err := workspaceInitCmd.RunE(workspaceInitCmd, nil); err != nil {
		t.Fatalf("workspaceInitCmd.RunE() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspace.Root(), "outputs", "artifact.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := workspaceStatusCmd.RunE(workspaceStatusCmd, nil); err != nil {
		t.Fatalf("workspaceStatusCmd.RunE() error = %v", err)
	}
	if err := workspaceBackupCmd.RunE(workspaceBackupCmd, nil); err != nil {
		t.Fatalf("workspaceBackupCmd.RunE() error = %v", err)
	}
	workspaceCleanOutputs = true
	if err := workspaceCleanCmd.RunE(workspaceCleanCmd, nil); err != nil {
		t.Fatalf("workspaceCleanCmd.RunE() error = %v", err)
	}
}

func TestInstallStartStopRestartAndExecCommands(t *testing.T) {
	restoreCommandDeps(t)
	buf := captureOutput(t)

	boosting, err := suite.Get("boosting")
	if err != nil {
		t.Fatalf("suite.Get(boosting) error = %v", err)
	}
	versions := config.Versions{}
	ensureWorkspaceLayout = func() error { return nil }
	loadVersions = func() (config.Versions, error) { return versions, nil }
	saveVersions = func(v config.Versions) error { versions = v; return nil }
	dockerPullWithProgress = func(ctx context.Context, image string, cb func(string)) error { return nil }
	dockerWriteSuiteCompose = func(ctx context.Context, s suite.Suite) (string, error) {
		return "/tmp/" + s.Name + ".compose.yml", nil
	}
	systemRegisterPorts = func(s suite.Suite) error { return nil }
	systemCheckConflicts = func(s suite.Suite) []system.PortConflict { return nil }
	if err := installCmd.RunE(installCmd, []string{"boosting"}); err != nil {
		t.Fatalf("installCmd.RunE(boosting) error = %v", err)
	}
	if _, ok := versions["boosting"]["gradient-boost-core"]; !ok {
		t.Fatalf("expected versions to be recorded, got %#v", versions)
	}

	selectForgeComponents = func() []string { return []string{"gradient-boost-core"} }
	buildForgeCompose = func(selected []string) ([]byte, error) { return []byte("services:\n"), nil }
	dockerWriteRawCompose = func(ctx context.Context, name string, data []byte) (string, error) {
		return "/tmp/" + name + ".compose.yml", nil
	}
	if err := installCmd.RunE(installCmd, []string{"forge"}); err != nil {
		t.Fatalf("installCmd.RunE(forge) error = %v", err)
	}

	loadState = func() (config.State, error) { return config.State{Installed: []string{"boosting", "flow"}}, nil }
	workspaceComposePath = func(name string) string { return "/tmp/" + name + ".compose.yml" }
	var upCalls []string
	var downCalls []string
	dockerComposeUp = func(ctx context.Context, path string, detach bool) error {
		upCalls = append(upCalls, path)
		return nil
	}
	dockerComposeDown = func(ctx context.Context, path string) error {
		downCalls = append(downCalls, path)
		return nil
	}
	if err := startCmd.RunE(startCmd, nil); err != nil {
		t.Fatalf("startCmd.RunE() error = %v", err)
	}
	if err := stopCmd.RunE(stopCmd, nil); err != nil {
		t.Fatalf("stopCmd.RunE() error = %v", err)
	}
	if err := restartCmd.RunE(restartCmd, []string{"boosting"}); err != nil {
		t.Fatalf("restartCmd.RunE() error = %v", err)
	}
	if len(upCalls) == 0 || len(downCalls) == 0 {
		t.Fatalf("expected compose calls, got up=%v down=%v", upCalls, downCalls)
	}

	getSuite = func(name string) (suite.Suite, error) { return boosting, nil }
	dockerExecCommand = func(ctx context.Context, container string, args ...string) error {
		if container != "gradient-boost-core" || len(args) != 2 {
			t.Fatalf("unexpected exec target %s %#v", container, args)
		}
		return nil
	}
	if err := execCmd.RunE(execCmd, []string{"boosting", "python", "-V"}); err != nil {
		t.Fatalf("execCmd.RunE() error = %v", err)
	}

	if !strings.Contains(buf.String(), "Installed") || !strings.Contains(buf.String(), "Started") {
		t.Fatalf("expected command output, got %q", buf.String())
	}
}

func TestUpdateRollbackRemoveListStatusAndChangelog(t *testing.T) {
	restoreCommandDeps(t)
	buf := captureOutput(t)

	boosting, err := suite.Get("boosting")
	if err != nil {
		t.Fatalf("suite.Get(boosting) error = %v", err)
	}
	getSuite = func(name string) (suite.Suite, error) { return boosting, nil }
	versions := config.Versions{}
	config.SetImageVersion(versions, "boosting", "gradient-boost-core", "python:3.11", "python:3.10")
	config.SetImageVersion(versions, "boosting", "gradient-boost-lab", "lab:new", "lab:old")
	config.SetImageVersion(versions, "boosting", "gradient-boost-track", "track:new", "track:old")
	loadVersions = func() (config.Versions, error) { return versions, nil }
	saveVersions = func(v config.Versions) error { versions = v; return nil }
	dockerPullWithProgress = func(ctx context.Context, image string, cb func(string)) error { return nil }
	dockerWriteSuiteCompose = func(ctx context.Context, s suite.Suite) (string, error) { return "/tmp/" + s.Name + ".yml", nil }
	dockerComposeDown = func(ctx context.Context, path string) error { return nil }
	dockerComposeUp = func(ctx context.Context, path string, detach bool) error { return nil }
	composeDir := t.TempDir()
	workspaceComposePath = func(name string) string { return filepath.Join(composeDir, name+".yml") }
	removeInstalledSuite = func(name string) error { return nil }
	systemDeregisterPorts = func(s suite.Suite) error { return nil }
	dockerContainerStatus = func(ctx context.Context, name string) (string, error) { return "running", nil }
	loadState = func() (config.State, error) { return config.State{Installed: []string{"boosting"}}, nil }

	if err := updateCmd.RunE(updateCmd, []string{"boosting"}); err != nil {
		t.Fatalf("updateCmd.RunE() error = %v", err)
	}
	if err := rollbackCmd.RunE(rollbackCmd, []string{"boosting"}); err != nil {
		t.Fatalf("rollbackCmd.RunE() error = %v", err)
	}
	if err := changelogCmd.RunE(changelogCmd, []string{"boosting"}); err != nil {
		t.Fatalf("changelogCmd.RunE() error = %v", err)
	}

	composePath := workspaceComposePath("boosting")
	if err := os.WriteFile(composePath, []byte("services:\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := removeCmd.RunE(removeCmd, []string{"boosting"}); err != nil {
		t.Fatalf("removeCmd.RunE() error = %v", err)
	}

	if err := listCmd.RunE(listCmd, nil); err != nil {
		t.Fatalf("listCmd.RunE() error = %v", err)
	}
	if err := statusCmd.RunE(statusCmd, nil); err != nil {
		t.Fatalf("statusCmd.RunE() error = %v", err)
	}

	for _, token := range []string{"Updated", "Rollback", "Removed", "installed suites", "current=", "running"} {
		if !strings.Contains(buf.String(), token) {
			t.Fatalf("expected %q in output %q", token, buf.String())
		}
	}
}

func TestLabLogsShellDriverWizardAndSetup(t *testing.T) {
	restoreCommandDeps(t)
	buf := captureOutput(t)

	boosting, err := suite.Get("boosting")
	if err != nil {
		t.Fatalf("suite.Get(boosting) error = %v", err)
	}
	loadState = func() (config.State, error) { return config.State{Installed: []string{"boosting"}}, nil }
	getSuite = func(name string) (suite.Suite, error) { return boosting, nil }
	runDockerOutput = func(ctx context.Context, args ...string) ([]byte, error) {
		if strings.Join(args, " ") == "exec gradient-boost-lab jupyter server list" {
			return []byte("http://0.0.0.0:8888/?token=abc123 :: /notebooks\n"), nil
		}
		return []byte("http://localhost:8888/?token=fallback\n"), nil
	}
	opened := ""
	systemOpenURL = func(url string) error { opened = url; return nil }
	if err := labCmd.RunE(labCmd, nil); err != nil {
		t.Fatalf("labCmd.RunE() error = %v", err)
	}
	if opened != "http://127.0.0.1:8888/lab?token=abc123" {
		t.Fatalf("unexpected opened URL %q", opened)
	}

	var interactiveCalls []string
	runDockerInteractive = func(ctx context.Context, args ...string) error {
		call := strings.Join(args, " ")
		interactiveCalls = append(interactiveCalls, call)
		if strings.Contains(call, " bash") {
			return errors.New("bash missing")
		}
		return nil
	}
	logsService = "gradient-boost-core"
	if err := logsCmd.RunE(logsCmd, []string{"boosting"}); err != nil {
		t.Fatalf("logsCmd.RunE() error = %v", err)
	}
	if err := shellCmd.RunE(shellCmd, []string{"boosting"}); err != nil {
		t.Fatalf("shellCmd.RunE() error = %v", err)
	}
	if len(interactiveCalls) < 3 {
		t.Fatalf("expected interactive docker calls, got %#v", interactiveCalls)
	}

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
	gpuDetectState = func() (gpu.GPUState, error) { return gpu.GPUStateNone, nil }
	installed := []string{}
	installCmd.RunE = func(cmd *cobra.Command, args []string) error {
		installed = append(installed, args[0])
		return nil
	}
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

	if !strings.Contains(buf.String(), "Secure Boot") || !strings.Contains(buf.String(), "GPU state") {
		t.Fatalf("expected wizard/setup output, got %q", buf.String())
	}
}

func TestSelfUpdateCommand(t *testing.T) {
	restoreCommandDeps(t)
	buf := captureOutput(t)

	binary := []byte("concave-binary")
	sum := sha256.Sum256(binary)
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/manifest":
			_ = json.NewEncoder(w).Encode(updateManifest{
				Version: "v0.1.0",
				URL:     server.URL + "/concave",
				SHA256:  hex.EncodeToString(sum[:]),
			})
		case "/concave":
			_, _ = w.Write(binary)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	selfUpdateClient = server.Client()
	selfUpdateManifestURL = server.URL + "/manifest"
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
