package cmd

import (
	"fmt"

	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var changelogCmd = &cobra.Command{
	Use:   "changelog [suite]",
	Short: "Show the diff between active image tags and registry tags",
	Args:  cobra.ExactArgs(1),
	RunE:  runChangelog,
}

func runChangelog(cmd *cobra.Command, args []string) error {
	name := args[0]
	s, err := currentSuiteDefinition(name)
	if err != nil {
		return err
	}
	manifest, err := loadManifest()
	if err != nil {
		return err
	}

	ui.Warn("Registry", "remote lookup is deferred; showing local target versions only")
	ui.Header("Gradient Linux — suite changelog")
	for _, container := range s.Containers {
		version, ok := manifest[s.Name][container.Name]
		if !ok {
			ui.Info(container.Name, "not installed yet")
			continue
		}
		if version.Current == container.Image {
			ui.Info(container.Name, "up to date")
			continue
		}
		ui.Info(container.Name, fmt.Sprintf("current=%s target=%s previous=%s", version.Current, container.Image, version.Previous))
	}
	return nil
}

func init() {
	rootCmd.AddCommand(changelogCmd)
}
