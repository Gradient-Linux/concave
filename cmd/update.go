package cmd

import (
	"context"
	"time"

	"github.com/gradient-linux/concave/internal/config"
	"github.com/gradient-linux/concave/internal/docker"
	"github.com/gradient-linux/concave/internal/suite"
	"github.com/gradient-linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [suite]",
	Short: "Update a suite to the images pinned in the registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := suite.Get(args[0])
		if err != nil {
			return err
		}

		versions, err := config.LoadVersions()
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		for _, container := range s.Containers {
			recorded, _ := config.GetImageVersion(versions, s.Name, container.Name)
			config.SetImageVersion(versions, s.Name, container.Name, container.Image, recorded.Current)
			ui.Info("Pulling", container.Image)
			if err := docker.PullWithProgress(ctx, container.Image, nil); err != nil {
				return err
			}
		}
		if err := config.SaveVersions(versions); err != nil {
			return err
		}
		if _, err := docker.WriteSuiteCompose(ctx, s); err != nil {
			return err
		}
		ui.Pass("Updated", s.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
