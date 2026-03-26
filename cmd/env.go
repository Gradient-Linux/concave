package cmd

import (
	"errors"
	"fmt"

	"github.com/Gradient-Linux/concave/internal/resolverclient"
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var (
	envGroup   string
	envLayers  string
	envBackend string
	envPackage string
)

const envLayerHelp = `Environment layers:
  Layer 1 (Hardware backend) — CUDA/ROCm version. Managed by concave, not exported.
  Layer 2 (Framework binaries) — torch+cu121, torch+rocm. Managed by concave, not exported.
  Layer 3 (Pure Python) — transformers, scikit-learn, pandas. Portable across hardware.

Use --layers python to export only Layer 3. This snapshot can be applied
to a different hardware backend with 'concave env apply --backend rocm'.`

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment baselines and drift",
}

var envStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show environment status",
	RunE:  runEnvStatus,
}

var envExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export environment snapshot",
	Long:  "Export an environment snapshot for a group.\n\n" + envLayerHelp,
	RunE:  runScaffoldCommand,
}

var envApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply environment snapshot",
	Long:  "Apply an environment snapshot for a group.\n\n" + envLayerHelp,
	RunE:  runScaffoldCommand,
}

var envDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff environment state against baseline",
	RunE:  runEnvDiff,
}

var envRollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback one package",
	RunE:  runScaffoldCommand,
}

var envBaselineCmd = &cobra.Command{
	Use:   "baseline",
	Short: "Manage environment baselines",
}

var envBaselineSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the current baseline",
	RunE:  runEnvBaselineSet,
}

var envBaselineShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current baseline",
	RunE:  runEnvBaselineShow,
}

func init() {
	envStatusCmd.Flags().StringVar(&envGroup, "group", "", "group name")
	envExportCmd.Flags().StringVar(&envGroup, "group", "", "group name")
	envExportCmd.Flags().StringVar(&envLayers, "layers", "python", "layers to export")
	envApplyCmd.Flags().StringVar(&envGroup, "group", "", "group name")
	envApplyCmd.Flags().StringVar(&envBackend, "backend", "cuda", "target backend (cuda, rocm, cpu)")
	envDiffCmd.Flags().StringVar(&envGroup, "group", "", "group name")
	envRollbackCmd.Flags().StringVar(&envPackage, "package", "", "package name")
	envRollbackCmd.Flags().StringVar(&envGroup, "group", "", "group name")
	envBaselineSetCmd.Flags().StringVar(&envGroup, "group", "", "group name")
	envBaselineShowCmd.Flags().StringVar(&envGroup, "group", "", "group name")

	envBaselineCmd.AddCommand(envBaselineSetCmd, envBaselineShowCmd)
	envCmd.AddCommand(envStatusCmd, envExportCmd, envApplyCmd, envDiffCmd, envRollbackCmd, envBaselineCmd)
	rootCmd.AddCommand(envCmd)
}

func runEnvStatus(cmd *cobra.Command, args []string) error {
	status, err := resolverclient.QueryStatus("")
	if errors.Is(err, resolverclient.ErrUnavailable) {
		ui.Warn("Resolver", maximaNotImplementedMessage)
		return nil
	}
	if err != nil {
		return err
	}

	if envGroup != "" {
		return printDriftReports(envGroup, status.GroupReports)
	}
	ui.Pass("Resolver", fmt.Sprintf("running · %d snapshots", status.SnapshotCount))
	if len(status.GroupReports) == 0 {
		ui.Info("Groups", "no baseline reports yet")
		return nil
	}
	return printDriftReports("", status.GroupReports)
}

func runEnvDiff(cmd *cobra.Command, args []string) error {
	reports, err := resolverclient.QueryDrift("", envGroup)
	if errors.Is(err, resolverclient.ErrUnavailable) {
		ui.Warn("Resolver", maximaNotImplementedMessage)
		return nil
	}
	if err != nil {
		return err
	}
	if len(reports) == 0 {
		ui.Info("Diff", "no drift reports available")
		return nil
	}
	for _, report := range reports {
		if report.Clean {
			ui.Pass(report.Group, "no package drift")
			continue
		}
		for _, diff := range report.Diffs {
			label := report.Group + "/" + diff.Name
			detail := fmt.Sprintf("%s -> %s (%s)", diff.Baseline, diff.Current, diff.Reason)
			ui.Info(label, detail)
		}
	}
	return nil
}

func printDriftReports(group string, reports []resolverclient.DriftReport) error {
	filtered := reports
	if group != "" {
		filtered = make([]resolverclient.DriftReport, 0, len(reports))
		for _, report := range reports {
			if report.Group == group {
				filtered = append(filtered, report)
			}
		}
	}
	if len(filtered) == 0 {
		ui.Info("Groups", "no drift reports available")
		return nil
	}
	for _, report := range filtered {
		if report.Clean {
			ui.Pass(report.Group, "clean")
			continue
		}
		ui.Warn(report.Group, fmt.Sprintf("%d packages flagged", len(report.Diffs)))
	}
	return nil
}

func runEnvBaselineSet(cmd *cobra.Command, args []string) error {
	snapshot, err := resolverclient.ApplyBaseline("", envGroup, "")
	if errors.Is(err, resolverclient.ErrUnavailable) {
		ui.Warn("Resolver", maximaNotImplementedMessage)
		return nil
	}
	if err != nil {
		return err
	}
	ui.Pass("Baseline", fmt.Sprintf("%s @ %s", snapshot.Group, snapshot.Timestamp.Format("2006-01-02 15:04:05 MST")))
	ui.Info("Packages", fmt.Sprintf("%d", len(snapshot.Packages)))
	return nil
}

func runEnvBaselineShow(cmd *cobra.Command, args []string) error {
	snapshot, err := resolverclient.QueryBaseline("", envGroup)
	if errors.Is(err, resolverclient.ErrUnavailable) {
		ui.Warn("Resolver", maximaNotImplementedMessage)
		return nil
	}
	if err != nil {
		return err
	}
	ui.Pass("Baseline", fmt.Sprintf("%s @ %s", snapshot.Group, snapshot.Timestamp.Format("2006-01-02 15:04:05 MST")))
	ui.Info("Packages", fmt.Sprintf("%d", len(snapshot.Packages)))
	return nil
}
