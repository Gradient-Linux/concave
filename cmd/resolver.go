package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/Gradient-Linux/concave/internal/resolverclient"
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var resolverFollow bool
var meshFollow bool

var resolverCmd = &cobra.Command{
	Use:   "resolver",
	Short: "Manage the concave-resolver service",
}

var resolverStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show resolver status",
	RunE:  runResolverStatus,
}

var resolverLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show resolver logs",
	RunE:  runScaffoldCommand,
}

var resolverRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the resolver service",
	RunE:  runScaffoldCommand,
}

var resolverRunCmd = &cobra.Command{
	Use:    "run",
	Short:  "Run the resolver daemon directly",
	Hidden: true,
	RunE:   runScaffoldCommand,
}

var meshCmd = &cobra.Command{
	Use:   "mesh",
	Short: "Manage the gradient-mesh service",
}

var meshStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show mesh status",
	RunE:  runNodeStatus,
}

var meshLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show mesh logs",
	RunE:  runScaffoldCommand,
}

var meshRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the mesh service",
	RunE:  runScaffoldCommand,
}

var meshRunCmd = &cobra.Command{
	Use:    "run",
	Short:  "Run the mesh daemon directly",
	Hidden: true,
	RunE:   runScaffoldCommand,
}

func init() {
	resolverLogsCmd.Flags().BoolVar(&resolverFollow, "follow", false, "follow resolver logs")
	meshLogsCmd.Flags().BoolVar(&meshFollow, "follow", false, "follow mesh logs")

	resolverCmd.AddCommand(resolverStatusCmd, resolverLogsCmd, resolverRestartCmd, resolverRunCmd)
	meshCmd.AddCommand(meshStatusCmd, meshLogsCmd, meshRestartCmd, meshRunCmd)
	rootCmd.AddCommand(resolverCmd, meshCmd)
}

func runResolverStatus(cmd *cobra.Command, args []string) error {
	status, err := resolverclient.QueryStatus("")
	if errors.Is(err, resolverclient.ErrUnavailable) {
		ui.Warn("Resolver", maximaNotImplementedMessage)
		return nil
	}
	if err != nil {
		return err
	}
	ui.Pass("Resolver", "running")
	if !status.LastScan.IsZero() {
		ui.Info("Last scan", status.LastScan.Format(time.RFC3339))
	}
	ui.Info("Snapshots", fmt.Sprintf("%d", status.SnapshotCount))
	for _, report := range status.GroupReports {
		if report.Clean {
			ui.Pass(report.Group, "clean")
		} else {
			ui.Warn(report.Group, fmt.Sprintf("%d packages flagged", len(report.Diffs)))
		}
	}
	return nil
}

func runResolverCheckSummary() {
	status, err := resolverclient.QueryStatus("")
	if errors.Is(err, resolverclient.ErrUnavailable) {
		ui.Warn("Resolver", "not configured")
		return
	}
	if err != nil {
		ui.Warn("Resolver", err.Error())
		return
	}
	if !status.Running {
		ui.Warn("Resolver", "socket reachable but daemon not running")
		return
	}
	ui.Pass("Resolver", fmt.Sprintf("%d snapshots", status.SnapshotCount))
}

func runComputeCheckSummary() {
	ui.Info("Compute", "not configured")
}
