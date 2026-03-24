package cmd

import (
	"fmt"
	"os/user"
	"strings"

	"github.com/Gradient-Linux/concave/internal/auth"
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the current Unix user, Gradient role, and allowed commands",
	RunE:  runWhoami,
}

func runWhoami(cmd *cobra.Command, args []string) error {
	current, err := user.Current()
	if err != nil {
		return err
	}

	ui.Info("User", current.Username)
	groups, err := current.GroupIds()
	if err == nil {
		names := make([]string, 0, len(groups))
		for _, gid := range groups {
			if group, lookupErr := lookupGroup(gid); lookupErr == nil {
				names = append(names, group.Name)
			}
		}
		ui.Info("Groups", strings.Join(names, ", "))
	}

	role, err := auth.ResolveRole(current.Username)
	if err != nil {
		ui.Warn("Role", "none — not authorised")
		ui.Info("Fix", fmt.Sprintf("Ask a sysadmin to run:\n  sudo usermod -aG gradient-viewer %s", current.Username))
		return nil
	}

	ui.Pass("Role", role.String())
	ui.Info("Can run", strings.Join(allowedCommandNames(role), ", "))
	cannot := disallowedCommandNames(role)
	if len(cannot) > 0 {
		ui.Info("Cannot", strings.Join(cannot, ", "))
	}
	return nil
}

func lookupGroup(gid string) (*user.Group, error) {
	return user.LookupGroupId(gid)
}

func allowedCommandNames(role auth.Role) []string {
	all := [][]string{
		{"check", "whoami", "gpu check", "gpu info"},
		{"status", "list", "logs", "changelog", "workspace status", "fleet status", "fleet peers", "node status", "env status", "env diff", "env baseline show", "resolver status", "resolver logs", "mesh status", "mesh logs", "team list", "team status"},
		{"lab", "shell", "exec"},
		{"install", "remove", "start", "stop", "restart", "update", "rollback", "workspace backup", "workspace prune", "team create", "team delete", "team add-user", "team remove-user", "node set", "env export", "env apply", "env rollback", "env baseline set"},
		{"serve", "gpu setup", "setup", "upgrade", "resolver restart", "mesh restart"},
	}
	allowed := make([]string, 0, 16)
	for idx, commands := range all {
		if idx == 0 || auth.Role(idx-1) <= role {
			allowed = append(allowed, commands...)
		}
	}
	return allowed
}

func disallowedCommandNames(role auth.Role) []string {
	all := []string{
		"status", "list", "logs", "changelog", "workspace status", "fleet status", "fleet peers", "node status", "env status", "env diff", "env baseline show", "resolver status", "resolver logs", "mesh status", "mesh logs", "team list", "team status",
		"lab", "shell", "exec",
		"install", "remove", "start", "stop", "restart", "update", "rollback", "workspace backup", "workspace prune", "team create", "team delete", "team add-user", "team remove-user", "node set", "env export", "env apply", "env rollback", "env baseline set",
		"serve", "gpu setup", "setup", "upgrade", "resolver restart", "mesh restart",
	}
	allowed := map[string]struct{}{}
	for _, command := range allowedCommandNames(role) {
		allowed[command] = struct{}{}
	}
	blocked := make([]string, 0, len(all))
	for _, command := range all {
		if _, ok := allowed[command]; !ok {
			blocked = append(blocked, command)
		}
	}
	return blocked
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
