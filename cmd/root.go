package cmd

import (
	"os"

	"github.com/gradientlinux/concave/internal/ui"
	"github.com/spf13/cobra"
)

// Version contains the current CLI version and is set by main.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:           "concave",
	Short:         "Manage Gradient Linux AI suites",
	Long:          "concave manages AI/ML Docker suites, workspace layout, GPU setup, and lifecycle tasks for Gradient Linux.",
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       Version,
}

// Execute runs the root command and exits on error.
func Execute() {
	rootCmd.Version = Version
	if err := rootCmd.Execute(); err != nil {
		ui.Fail("Error", err.Error())
		os.Exit(1)
	}
}
