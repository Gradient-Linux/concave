package cmd

import (
	"context"
	"time"

	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/Gradient-Linux/concave/internal/ui"
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

		versions, err := loadVersions()
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		for _, container := range s.Containers {
			recorded, _ := getImageVersion(versions, s.Name, container.Name)
			setImageVersion(versions, s.Name, container.Name, container.Image, recorded.Current)
			ui.Info("Pulling", container.Image)
			if err := dockerPullWithProgress(ctx, container.Image, nil); err != nil {
				return err
			}
		}
		if err := saveVersions(versions); err != nil {
			return err
		}
		if _, err := dockerWriteSuiteCompose(ctx, s); err != nil {
			return err
		}
		ui.Pass("Updated", s.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
