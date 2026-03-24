package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Gradient-Linux/concave/internal/meshclient"
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

type NodeVisibility string

const (
	VisibilityPublic  NodeVisibility = "public"
	VisibilityPrivate NodeVisibility = "private"
	VisibilityHidden  NodeVisibility = "hidden"
)

var nodeVisibility string

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage this Gradient Linux node",
}

var nodeStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show node status",
	RunE:  runNodeStatus,
}

var nodeSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set node visibility",
	RunE:  runScaffoldCommand,
}

var fleetCmd = &cobra.Command{
	Use:   "fleet",
	Short: "Inspect fleet peers and node discovery",
}

var fleetStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show fleet status",
	RunE:  runFleetStatus,
}

var fleetPeersCmd = &cobra.Command{
	Use:   "peers",
	Short: "Show fleet peers",
	RunE:  runFleetPeers,
}

func init() {
	nodeSetCmd.Flags().StringVar(&nodeVisibility, "visibility", string(VisibilityPublic), "node visibility (public, private, hidden)")
	nodeCmd.AddCommand(nodeStatusCmd, nodeSetCmd)
	fleetCmd.AddCommand(fleetStatusCmd, fleetPeersCmd)
	rootCmd.AddCommand(nodeCmd, fleetCmd)
}

func runNodeStatus(cmd *cobra.Command, args []string) error {
	node, err := meshclient.QuerySelf("")
	if errors.Is(err, meshclient.ErrUnavailable) {
		ui.Warn("Mesh", maximaNotImplementedMessage)
		return nil
	}
	if err != nil {
		return err
	}
	ui.Pass("Node", node.Hostname)
	ui.Info("Visibility", string(node.Visibility))
	ui.Info("Resolver", fmt.Sprintf("%t", node.ResolverRunning))
	if len(node.InstalledSuites) > 0 {
		ui.Info("Suites", strings.Join(node.InstalledSuites, ", "))
	}
	return nil
}

func runFleetStatus(cmd *cobra.Command, args []string) error {
	peers, err := meshclient.QueryFleet("")
	if errors.Is(err, meshclient.ErrUnavailable) {
		ui.Warn("Mesh", maximaNotImplementedMessage)
		return nil
	}
	if err != nil {
		return err
	}
	if len(peers) == 0 {
		ui.Info("Fleet", "no peers visible")
		return nil
	}
	ui.Pass("Fleet", fmt.Sprintf("%d visible peers", len(peers)))
	for _, peer := range peers {
		ui.Info(peer.Hostname, fmt.Sprintf("%s · seen %s", peer.Visibility, peer.LastSeen.Format(time.RFC3339)))
	}
	return nil
}

func runFleetPeers(cmd *cobra.Command, args []string) error {
	return runFleetStatus(cmd, args)
}

func runMeshCheckSummary() {
	peers, err := meshclient.QueryFleet("")
	if errors.Is(err, meshclient.ErrUnavailable) {
		ui.Warn("Mesh", "not configured")
		return
	}
	if err != nil {
		ui.Warn("Mesh", err.Error())
		return
	}
	ui.Pass("Mesh", fmt.Sprintf("%d visible peers", len(peers)))
}
