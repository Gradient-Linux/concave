package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/gradient-linux/concave/internal/config"
	"github.com/gradient-linux/concave/internal/docker"
	"github.com/gradient-linux/concave/internal/suite"
	"github.com/gradient-linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show container status for installed suites",
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := config.LoadState()
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		ui.Header("Gradient Linux — suite status")
		for _, name := range state.Installed {
			s, err := suite.Get(name)
			if err != nil {
				return err
			}
			for _, container := range s.Containers {
				status, err := docker.ContainerStatus(ctx, container.Name)
				if err != nil {
					ui.Warn(container.Name, err.Error())
					continue
				}
				ui.Info(container.Name, fmt.Sprintf("%s (%s)", status, container.Role))
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
