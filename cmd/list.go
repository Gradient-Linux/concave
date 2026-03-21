package cmd

import (
	"fmt"

	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed suites",
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := loadState()
		if err != nil {
			return err
		}

		ui.Header("Gradient Linux — installed suites")
		for _, name := range state.Installed {
			s, err := getSuite(name)
			if err != nil {
				return err
			}
			ui.Info(name, fmt.Sprintf("containers=%d ports=%s", len(s.Containers), suitePorts(s)))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
