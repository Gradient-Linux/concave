package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/Gradient-Linux/concave/internal/gpu"
	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [suite]",
	Short: "Install a suite",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := ensureWorkspaceLayout(); err != nil {
			return err
		}

		var target suite.Suite
		if name == "forge" {
			selected := selectForgeComponents()
			if len(selected) == 0 {
				return fmt.Errorf("forge requires at least one selected component")
			}
			target = suite.Suite{Name: "forge", ComposeTemplate: "forge"}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			data, err := buildForgeCompose(selected)
			if err != nil {
				return err
			}
			path, err := dockerWriteRawCompose(ctx, "forge", data)
			if err != nil {
				return err
			}
			if err := addInstalledSuite("forge"); err != nil {
				return err
			}
			ui.Pass("Installed", path)
			return nil
		}

		plan, err := buildInstallPlan(name)
		if err != nil {
			return err
		}
		target = plan.Suite

		if target.GPURequired {
			state, err := gpuDetectState()
			if err != nil {
				return err
			}
			if state == gpu.GPUStateNone {
				ui.Warn("GPU", "Neural suite requires NVIDIA support for the full runtime path")
			}
		}

		if conflicts := systemCheckConflicts(target); len(conflicts) > 0 {
			conflict := conflicts[0]
			return fmt.Errorf("port %d conflicts with %s (%s)", conflict.Port, conflict.ExistingSuite, conflict.ExistingService)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		versions, err := loadVersions()
		if err != nil {
			return err
		}
		for _, container := range target.Containers {
			ui.Info("Pulling", container.Image)
			if err := dockerPullWithProgress(ctx, container.Image, nil); err != nil {
				return err
			}
			recorded, ok := getImageVersion(versions, target.Name, container.Name)
			previous := ""
			if ok {
				previous = recorded.Current
			}
			setImageVersion(versions, target.Name, container.Name, container.Image, previous)
		}
		if err := saveVersions(versions); err != nil {
			return err
		}

		path, err := dockerWriteSuiteCompose(ctx, target)
		if err != nil {
			return err
		}
		if err := addInstalledSuite(target.Name); err != nil {
			return err
		}
		if err := systemRegisterPorts(target); err != nil {
			return err
		}

		ui.Pass("Installed", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
