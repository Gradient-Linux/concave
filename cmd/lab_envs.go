package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/Gradient-Linux/concave/internal/lab"
	"github.com/Gradient-Linux/concave/internal/ui"
)

var (
	labLaunchImage       string
	labLaunchDisplayName string
	labLaunchTTL         time.Duration
	labLaunchGPUs        int
	labLaunchCPU         string
	labLaunchMem         string
	labLaunchDriver      string
	labLaunchOwner       string

	labExtendBy     time.Duration
	labStorageHot   string
	labStorageCold  string
	labActiveDriver string
)

var labEnvsCmd = &cobra.Command{
	Use:   "envs",
	Short: "Manage ephemeral JupyterLab environments",
}

var labEnvsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ephemeral lab environments",
	RunE:  runLabEnvsList,
}

var labEnvsLaunchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch an ephemeral lab environment",
	RunE:  runLabEnvsLaunch,
}

var labEnvsExtendCmd = &cobra.Command{
	Use:   "extend <env-id>",
	Short: "Extend the TTL of an ephemeral lab environment",
	Args:  cobra.ExactArgs(1),
	RunE:  runLabEnvsExtend,
}

var labEnvsArchiveCmd = &cobra.Command{
	Use:   "archive <env-id>",
	Short: "Archive an ephemeral lab environment to the cold tier and destroy it",
	Args:  cobra.ExactArgs(1),
	RunE:  runLabEnvsArchive,
}

var labStorageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Inspect or update lab storage tiers",
}

var labStorageShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current lab storage configuration",
	RunE:  runLabStorageShow,
}

var labStorageSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set hot and cold tier paths",
	RunE:  runLabStorageSet,
}

var labDriverCmd = &cobra.Command{
	Use:   "driver",
	Short: "Inspect or select the active lab driver",
}

var labDriverShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show registered lab drivers and the active one",
	RunE:  runLabDriverShow,
}

var labDriverSetCmd = &cobra.Command{
	Use:   "set <driver>",
	Short: "Set the active lab driver (docker | slurm | proxmox)",
	Args:  cobra.ExactArgs(1),
	RunE:  runLabDriverSet,
}

func runLabEnvsList(_ *cobra.Command, _ []string) error {
	mgr, err := newLabManager()
	if err != nil {
		return err
	}
	envs := mgr.List()
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tOWNER\tDRIVER\tSTATUS\tIMAGE\tEXPIRES\tURL")
	now := time.Now()
	for _, env := range envs {
		expires := "-"
		if !env.ExpiresAt.IsZero() {
			expires = env.Remaining(now).Round(time.Second).String()
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			env.ID, env.Owner, env.Driver, env.Status, env.Image, expires, env.JupyterURL)
	}
	return tw.Flush()
}

func runLabEnvsLaunch(cmd *cobra.Command, _ []string) error {
	if labLaunchImage == "" {
		return fmt.Errorf("--image is required")
	}
	if labLaunchTTL <= 0 {
		return fmt.Errorf("--ttl must be positive (e.g. 2h)")
	}
	owner := labLaunchOwner
	if owner == "" {
		if u := os.Getenv("USER"); u != "" {
			owner = u
		} else {
			owner = "local"
		}
	}
	mgr, err := newLabManager()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
	defer cancel()
	env, err := mgr.Launch(ctx, lab.EnvSpec{
		Owner:       owner,
		Image:       labLaunchImage,
		DisplayName: labLaunchDisplayName,
		GPUs:        labLaunchGPUs,
		CPURequest:  labLaunchCPU,
		MemRequest:  labLaunchMem,
		TTL:         labLaunchTTL,
		Driver:      labLaunchDriver,
	})
	if err != nil {
		return err
	}
	ui.Info("lab", fmt.Sprintf("launched %s (%s) expiring in %s", env.ID, env.Driver, labLaunchTTL))
	if env.JupyterURL != "" {
		ui.Info("jupyter", env.JupyterURL)
	}
	return nil
}

func runLabEnvsExtend(cmd *cobra.Command, args []string) error {
	if labExtendBy <= 0 {
		return fmt.Errorf("--by must be positive (e.g. 1h)")
	}
	mgr, err := newLabManager()
	if err != nil {
		return err
	}
	env, err := mgr.ExtendTTL(cmd.Context(), args[0], labExtendBy)
	if err != nil {
		return err
	}
	ui.Info("lab", fmt.Sprintf("extended %s; new expiry %s", env.ID, env.ExpiresAt.Format(time.RFC3339)))
	return nil
}

func runLabEnvsArchive(cmd *cobra.Command, args []string) error {
	mgr, err := newLabManager()
	if err != nil {
		return err
	}
	env, err := mgr.ArchiveAndDestroy(cmd.Context(), args[0])
	if err != nil {
		return err
	}
	ui.Info("lab", fmt.Sprintf("archived %s -> %s", env.ID, env.ArchiveRef))
	return nil
}

func runLabStorageShow(_ *cobra.Command, _ []string) error {
	storage, err := lab.LoadStorage()
	if err != nil {
		return err
	}
	out, _ := json.MarshalIndent(storage, "", "  ")
	fmt.Println(string(out))
	return nil
}

func runLabStorageSet(_ *cobra.Command, _ []string) error {
	current, err := lab.LoadStorage()
	if err != nil {
		return err
	}
	if strings.TrimSpace(labStorageHot) != "" {
		current.HotTier = labStorageHot
	}
	if strings.TrimSpace(labStorageCold) != "" {
		current.ColdTier = labStorageCold
	}
	if err := lab.SaveStorage(current); err != nil {
		return err
	}
	if err := lab.EnsureTierDirs(current); err != nil {
		return err
	}
	ui.Info("lab", fmt.Sprintf("hot=%s cold=%s", current.HotTier, current.ColdTier))
	return nil
}

func runLabDriverShow(_ *cobra.Command, _ []string) error {
	mgr, err := newLabManager()
	if err != nil {
		return err
	}
	fmt.Println("active:", mgr.Registry().Active())
	fmt.Println("drivers:", strings.Join(mgr.Registry().Names(), ", "))
	return nil
}

func runLabDriverSet(_ *cobra.Command, args []string) error {
	mgr, err := newLabManager()
	if err != nil {
		return err
	}
	if err := mgr.Registry().SetActive(args[0]); err != nil {
		return err
	}
	ui.Info("lab", "active driver: "+args[0])
	return nil
}

// newLabManager builds a local-only lab.Manager suitable for CLI usage. The
// API server builds its own instance.
func newLabManager() (*lab.Manager, error) {
	registry := lab.NewRegistry()
	registry.Register(lab.NewDockerDriver(lab.DockerDriverConfig{}))
	store, err := lab.NewStore()
	if err != nil {
		return nil, err
	}
	storage, err := lab.LoadStorage()
	if err != nil {
		return nil, err
	}
	if err := lab.EnsureTierDirs(storage); err != nil {
		return nil, err
	}
	return lab.NewManager(registry, store, storage), nil
}

func init() {
	labEnvsLaunchCmd.Flags().StringVar(&labLaunchImage, "image", "", "docker image for the env (e.g. jupyter/datascience-notebook:latest)")
	labEnvsLaunchCmd.Flags().StringVar(&labLaunchDisplayName, "name", "", "display name")
	labEnvsLaunchCmd.Flags().DurationVar(&labLaunchTTL, "ttl", 2*time.Hour, "time-to-live before the env is archived")
	labEnvsLaunchCmd.Flags().IntVar(&labLaunchGPUs, "gpus", 0, "number of GPUs to attach")
	labEnvsLaunchCmd.Flags().StringVar(&labLaunchCPU, "cpu", "", "CPU request (docker --cpus)")
	labEnvsLaunchCmd.Flags().StringVar(&labLaunchMem, "mem", "", "memory request (docker --memory, e.g. 8g)")
	labEnvsLaunchCmd.Flags().StringVar(&labLaunchDriver, "driver", "", "driver to use (defaults to active)")
	labEnvsLaunchCmd.Flags().StringVar(&labLaunchOwner, "owner", "", "override owner (defaults to $USER)")

	labEnvsExtendCmd.Flags().DurationVar(&labExtendBy, "by", time.Hour, "extension duration")

	labStorageSetCmd.Flags().StringVar(&labStorageHot, "hot", "", "absolute path for the hot tier")
	labStorageSetCmd.Flags().StringVar(&labStorageCold, "cold", "", "absolute path for the cold tier")
	_ = labActiveDriver // reserved for future use

	labEnvsCmd.AddCommand(labEnvsListCmd, labEnvsLaunchCmd, labEnvsExtendCmd, labEnvsArchiveCmd)
	labStorageCmd.AddCommand(labStorageShowCmd, labStorageSetCmd)
	labDriverCmd.AddCommand(labDriverShowCmd, labDriverSetCmd)
	labCmd.AddCommand(labEnvsCmd, labStorageCmd, labDriverCmd)
}
