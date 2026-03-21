package cmd

import (
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

type exitCoder interface {
	ExitCode() int
}

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
		if code, ok := resolveExitCode(err); ok && code >= 0 {
			exitFunc(code)
			return
		}
		exitFunc(1)
	}
}

func resolveExitCode(err error) (int, bool) {
	for current := err; current != nil; {
		if coded, ok := current.(exitCoder); ok {
			return coded.ExitCode(), true
		}
		unwrapper, ok := current.(interface{ Unwrap() error })
		if !ok {
			break
		}
		current = unwrapper.Unwrap()
	}
	return 0, false
}
