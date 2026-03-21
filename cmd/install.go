package cmd

import (
	"context"
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
	state, err := gpuDetectState()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	return installSuite(ctx, args[0], suite.InstallOptions{
		GPUAvailable: state == gpu.GPUStateNVIDIA,
		Force:        installForce,
	})
}

func init() {
	installCmd.Flags().BoolVar(&installForce, "force", false, "reinstall an already-installed suite")
	rootCmd.AddCommand(installCmd)
}
