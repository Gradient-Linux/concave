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
	"github.com/Gradient-Linux/concave/internal/workspace"
)

var (
	exitFunc = os.Exit

	ensureWorkspaceLayout = workspace.EnsureLayout
	workspaceExists       = workspace.Exists
	workspaceRoot         = workspace.Root
	workspaceStatus       = workspace.Status
	workspaceBackup       = workspace.Backup
	workspaceClean        = workspace.CleanOutputs
	workspaceComposePath  = workspace.ComposePath

	loadState            = config.LoadState
	addInstalledSuite    = config.AddInstalled
	removeInstalledSuite = config.RemoveInstalled
	loadVersions         = config.LoadVersions
	saveVersions         = config.SaveVersions
	getImageVersion      = config.GetImageVersion
	setImageVersion      = config.SetImageVersion
	removeSuiteVersions  = config.RemoveSuiteVersions
	swapPreviousVersions = config.SwapPrevious

	getSuite              = suite.Get
	buildInstallPlan      = suite.BuildInstallPlan
	selectForgeComponents = suite.SelectForgeComponents
	buildForgeCompose     = suite.BuildForgeCompose
	suiteNames            = suite.Names
	primaryContainer      = suite.PrimaryContainer
	jupyterContainer      = suite.JupyterContainer
	suitePorts            = suite.SuitePorts

	dockerPullWithProgress  = docker.PullWithProgress
	dockerWriteSuiteCompose = docker.WriteSuiteCompose
	dockerWriteRawCompose   = docker.WriteRawCompose
	dockerComposeUp         = docker.ComposeUp
	dockerComposeDown       = docker.ComposeDown
	dockerContainerStatus   = docker.ContainerStatus
	dockerExecCommand       = docker.Exec

	gpuDetectState             = gpu.Detect
	gpuDetectAMDState          = gpu.DetectAMD
	gpuRecommendedDriverBranch = gpu.RecommendedDriverBranch
	gpuToolkitConfigured       = gpu.ToolkitConfigured
	gpuVerifyPassthrough       = gpu.VerifyPassthrough
	gpuSecureBootEnabled       = gpu.SecureBootEnabled

	systemDockerRunning     = system.DockerRunning
	systemUserInDockerGroup = system.UserInDockerGroup
	systemInternetReachable = system.InternetReachable
	systemCheckConflicts    = system.CheckConflicts
	systemRegisterPorts     = system.Register
	systemDeregisterPorts   = system.Deregister
	systemOpenURL           = system.OpenURL

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
