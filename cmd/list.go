package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all suites and their current state",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	state, err := loadState()
	if err != nil {
		return err
	}
	manifest, err := loadManifest()
	if err != nil {
		return err
	}

	installed := make(map[string]struct{}, len(state.Installed))
	for _, name := range state.Installed {
		installed[name] = struct{}{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	ui.Line("Installed Suites")
	ui.Line("─────────────────────────────────────────────────────")
	ui.Line(fmt.Sprintf("%-10s %-32s %s", "Suite", "Version", "Status"))
	ui.Line("─────────────────────────────────────────────────────")
	for _, name := range suiteNames() {
		version := "not installed"
		status := "—"
		if _, ok := installed[name]; ok {
			version = currentImageForFirstContainer(name, manifest)
			s, err := currentSuiteDefinition(name)
			if err != nil {
				return err
			}
			status, err = composePrimaryStatus(ctx, s)
			if err != nil {
				status = "error"
			}
		}
		ui.Line(fmt.Sprintf("%-10s %-32s %s", name, version, status))
	}
	ui.Line("─────────────────────────────────────────────────────")
	return nil
}

func init() {
	rootCmd.AddCommand(listCmd)
}
