package cmd

import (
	"context"
	"os"
	"os/exec"

	"github.com/Gradient-Linux/concave/internal/config"
	"github.com/Gradient-Linux/concave/internal/docker"
	"github.com/Gradient-Linux/concave/internal/gpu"
	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/Gradient-Linux/concave/internal/system"
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/Gradient-Linux/concave/internal/workspace"
)

var (
	exitFunc = os.Exit

	ensureWorkspaceLayout = workspace.EnsureLayout
	workspaceExists       = workspace.Exists
	workspaceRoot         = workspace.Root
	workspaceUserRoot     = workspace.UserRoot
	workspaceStatus       = workspace.Status
	workspaceBackup       = workspace.Backup
	workspaceClean        = workspace.CleanOutputs

	loadState         = config.LoadState
	saveState         = config.SaveState
	addSuite          = config.AddSuite
	removeSuite       = config.RemoveSuite
	isInstalled       = config.IsInstalled
	loadManifest      = config.LoadManifest
	saveManifest      = config.SaveManifest
	loadSetupState    = config.LoadSetupState
	saveSetupState    = config.SaveSetupState
	markStepComplete  = config.MarkStepComplete
	markSetupComplete = config.MarkSetupComplete
	isStepComplete    = config.IsStepComplete
	recordInstall     = config.RecordInstall
	recordUpdate      = config.RecordUpdate
	swapForRollback   = config.SwapForRollback

	getSuite                = suite.Get
	allSuites               = suite.All
	suiteNames              = suite.Names
	primaryContainer        = suite.PrimaryContainer
	jupyterContainer        = suite.JupyterContainer
	pickForgeComponents     = suite.PickComponents
	buildForgeCompose       = suite.BuildForgeCompose
	forgeSelectionFromNames = suite.SelectionFromContainerNames
	installSuite            = suite.Install
	waitHealthy             = suite.WaitHealthy

	dockerPullWithRollbackSafety = docker.PullWithRollbackSafety
	dockerWriteCompose           = docker.WriteCompose
	dockerWriteRawCompose        = docker.WriteRawCompose
	dockerComposePath            = docker.ComposePath
	dockerComposeUp              = docker.ComposeUp
	dockerComposeDown            = docker.ComposeDown
	dockerContainerStatus        = docker.ContainerStatus
	dockerComposeServiceStatus   = docker.ComposeServiceStatus
	dockerComposeExecOutput      = docker.ComposeExecOutput
	dockerComposeExecInteractive = docker.ComposeExecInteractive
	dockerContainerLogs          = docker.ContainerLogs
	dockerExecCommand            = docker.Exec
	dockerRevertToPrevious       = docker.RevertToPrevious

	gpuDetectState             = gpu.Detect
	gpuDetectAMDState          = gpu.DetectAMD
	gpuComputeCapability       = gpu.ComputeCapability
	gpuNVIDIADevices           = gpu.NVIDIADevices
	gpuRecommendedDriverBranch = gpu.RecommendedDriverBranch
	gpuToolkitConfigured       = gpu.ToolkitConfigured
	gpuVerifyPassthrough       = gpu.VerifyPassthrough
	gpuSecureBootEnabled       = gpu.SecureBootEnabled

	systemDockerRunning     = system.DockerRunning
	systemCommandAvailable  = system.CommandAvailable
	systemDockerCompose     = system.DockerComposeAvailable
	systemUserInDockerGroup = system.UserInDockerGroup
	systemInternetReachable = system.InternetReachable
	systemRunPrivileged     = system.RunPrivileged
	systemCheckConflicts    = system.CheckConflicts
	systemRegisterPorts     = system.Register
	systemDeregisterPorts   = system.Deregister
	systemOpenURL           = system.OpenURL
	systemSignalHandler     = system.WithSignalHandler
	systemLock              = system.Lock
	uiConfirm               = ui.Confirm
	currentUsername         = func() string {
		if value := os.Getenv("SUDO_USER"); value != "" {
			return value
		}
		if value := os.Getenv("USER"); value != "" {
			return value
		}
		return "unknown"
	}

	runDockerOutput = func(ctx context.Context, args ...string) ([]byte, error) {
		return exec.CommandContext(ctx, "docker", args...).CombinedOutput()
	}
	runDockerInteractive = func(ctx context.Context, args ...string) error {
		command := exec.CommandContext(ctx, "docker", args...)
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		return command.Run()
	}
)

func init() {
	suite.SetConflictChecker(func(s suite.Suite, installed []string) ([]suite.PortConflict, error) {
		conflicts, err := systemCheckConflicts(s, installed)
		if err != nil {
			return nil, err
		}
		mapped := make([]suite.PortConflict, 0, len(conflicts))
		for _, conflict := range conflicts {
			mapped = append(mapped, suite.PortConflict{
				Port:          conflict.Port,
				ExistingSuite: conflict.ExistingSuite,
				NewSuite:      conflict.NewSuite,
				Service:       conflict.Service,
			})
		}
		return mapped, nil
	})
}
