package cmd

import (
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

const maximaNotImplementedMessage = "not yet implemented — available in Gradient Linux Maxima after concave-resolver and compute engine are configured"

func runScaffoldCommand(cmd *cobra.Command, args []string) error {
	ui.Info(cmd.CommandPath(), maximaNotImplementedMessage)
	return nil
}
