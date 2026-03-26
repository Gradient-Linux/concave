package cmd

import (
	"context"
	"strings"
	"time"

	"github.com/Gradient-Linux/concave/internal/gpu"
	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/spf13/cobra"
)

var installForce bool

var installCmd = &cobra.Command{
	Use:   "install [suite]",
	Short: "Install a Gradient Linux AI suite",
	Long:  "Install one of: neural, boosting, flow, forge",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstall,
}

func runInstall(cmd *cobra.Command, args []string) error {
	if _, err := getSuite(args[0]); err != nil {
		return err
	}

	state, err := gpuDetectState()
	if err != nil {
		return err
	}

	return runLockedOperation("install", 45*time.Minute, composeCleanup(args[0]), func(ctx context.Context) error {
		if err := ensureDockerRuntime(ctx, "install "+args[0]); err != nil {
			return err
		}
		err := installSuite(ctx, args[0], suite.InstallOptions{
			GPUAvailable: state == gpu.GPUStateNVIDIA,
			Force:        installForce,
		})
		if err != nil && (strings.HasPrefix(err.Error(), "step 5 pull images") || strings.HasPrefix(err.Error(), "step 6 write compose file")) {
			return wrapDockerError(err)
		}
		return err
	})
}

func init() {
	installCmd.Flags().BoolVar(&installForce, "force", false, "reinstall an already-installed suite")
	rootCmd.AddCommand(installCmd)
}
