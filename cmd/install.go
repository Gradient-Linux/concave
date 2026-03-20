package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/gradientlinux/concave/internal/config"
	"github.com/gradientlinux/concave/internal/docker"
	"github.com/gradientlinux/concave/internal/gpu"
	"github.com/gradientlinux/concave/internal/suite"
	"github.com/gradientlinux/concave/internal/system"
	"github.com/gradientlinux/concave/internal/ui"
	"github.com/gradientlinux/concave/internal/workspace"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [suite]",
	Short: "Install a suite",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := workspace.EnsureLayout(); err != nil {
			return err
		}

		var target suite.Suite
		if name == "forge" {
			selected := suite.SelectForgeComponents()
			if len(selected) == 0 {
				return fmt.Errorf("forge requires at least one selected component")
			}
			target = suite.Suite{Name: "forge", ComposeTemplate: "forge"}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			data, err := suite.BuildForgeCompose(selected)
			if err != nil {
				return err
			}
			path, err := docker.WriteRawCompose(ctx, "forge", data)
			if err != nil {
				return err
			}
			if err := config.AddInstalled("forge"); err != nil {
				return err
			}
			ui.Pass("Installed", path)
			return nil
		}

		plan, err := suite.BuildInstallPlan(name)
		if err != nil {
			return err
		}
		target = plan.Suite

		if target.GPURequired {
			state, err := gpu.Detect()
			if err != nil {
				return err
			}
			if state == gpu.GPUStateNone {
				ui.Warn("GPU", "Neural suite requires NVIDIA support for the full runtime path")
			}
		}

		if conflicts := system.CheckConflicts(target); len(conflicts) > 0 {
			conflict := conflicts[0]
			return fmt.Errorf("port %d conflicts with %s (%s)", conflict.Port, conflict.ExistingSuite, conflict.ExistingService)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		versions, err := config.LoadVersions()
		if err != nil {
			return err
		}
		for _, container := range target.Containers {
			ui.Info("Pulling", container.Image)
			if err := docker.PullWithProgress(ctx, container.Image, nil); err != nil {
				return err
			}
			recorded, ok := config.GetImageVersion(versions, target.Name, container.Name)
			previous := ""
			if ok {
				previous = recorded.Current
			}
			config.SetImageVersion(versions, target.Name, container.Name, container.Image, previous)
		}
		if err := config.SaveVersions(versions); err != nil {
			return err
		}

		path, err := docker.WriteSuiteCompose(ctx, target)
		if err != nil {
			return err
		}
		if err := config.AddInstalled(target.Name); err != nil {
			return err
		}
		if err := system.Register(target); err != nil {
			return err
		}

		ui.Pass("Installed", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
