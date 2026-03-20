package cmd

import (
	"fmt"

	"github.com/gradient-linux/concave/internal/config"
	"github.com/gradient-linux/concave/internal/suite"
	"github.com/gradient-linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var changelogCmd = &cobra.Command{
	Use:   "changelog [suite]",
	Short: "Show the diff between active image tags and registry tags",
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

		ui.Header("Gradient Linux — suite changelog")
		for _, container := range s.Containers {
			version, ok := config.GetImageVersion(versions, s.Name, container.Name)
			if !ok {
				ui.Info(container.Name, "not installed yet")
				continue
			}
			ui.Info(container.Name, fmt.Sprintf("current=%s target=%s previous=%s", version.Current, container.Image, version.Previous))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(changelogCmd)
}
