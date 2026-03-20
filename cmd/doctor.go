package cmd

import (
	"fmt"
	"os"

	"github.com/gradient-linux/concave/internal/system"
	"github.com/gradient-linux/concave/internal/ui"
	"github.com/gradient-linux/concave/internal/workspace"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run concave system health checks",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Header("Gradient Linux — concave doctor")

		if ok, err := system.DockerRunning(); err != nil {
			ui.Fail("Docker", err.Error())
		} else if ok {
			ui.Pass("Docker", "running")
		} else {
			ui.Fail("Docker", "not running")
		}

		if ok, err := system.UserInDockerGroup(); err != nil {
			ui.Fail("Docker group", err.Error())
		} else if ok {
			ui.Pass("Docker group", "membership detected")
		} else {
			ui.Warn("Docker group", "user not in docker group")
		}

		if ok, err := system.InternetReachable(); err != nil {
			ui.Fail("Internet", err.Error())
		} else if ok {
			ui.Pass("Internet", "reachable")
		} else {
			ui.Warn("Internet", "not reachable")
		}

		if workspace.Exists() {
			ui.Pass("Workspace", workspace.Root())
		} else {
			ui.Warn("Workspace", "not initialized")
		}

		// GPU_SECTION_START — GPU Agent adds checks here in Phase 4
		// GPU_SECTION_END

		if _, err := os.Stat(workspace.Root()); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("workspace stat %s: %w", workspace.Root(), err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
