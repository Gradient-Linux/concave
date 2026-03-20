package cmd

import (
	"fmt"

	"github.com/gradientlinux/concave/internal/config"
	"github.com/gradientlinux/concave/internal/suite"
	"github.com/gradientlinux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed suites",
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := config.LoadState()
		if err != nil {
			return err
		}

		ui.Header("Gradient Linux — installed suites")
		for _, name := range state.Installed {
			s, err := suite.Get(name)
			if err != nil {
				return err
			}
			ui.Info(name, fmt.Sprintf("containers=%d ports=%s", len(s.Containers), suite.SuitePorts(s)))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
