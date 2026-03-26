package cmd

import (
	"os"
	"strings"

	"github.com/Gradient-Linux/concave/internal/auth"
	"github.com/Gradient-Linux/concave/internal/system"
	"github.com/Gradient-Linux/concave/internal/ui"
	"github.com/spf13/cobra"
)

type exitCoder interface {
	ExitCode() int
}

// Version contains the current CLI version and is set by main.
var Version = "dev"
var Commit = "none"
var BuildDate = "unknown"
var verbose bool

var rootCmd = &cobra.Command{
	Use:           "concave",
	Short:         "Manage Gradient Linux AI suites",
	Long:          "concave manages AI/ML Docker suites, workspace layout, GPU setup, and lifecycle tasks for Gradient Linux.",
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		configureDefaultWorkspaceRoot(cmd)
		if isExemptCommand(cmd) {
			return nil
		}
		if _, err := auth.InitCLIRole(); err != nil {
			ui.Fail("Auth", err.Error())
			ui.Info("Fix", "Ask a sysadmin to run: usermod -aG gradient-viewer "+os.Getenv("USER"))
			return system.NewExitError(system.ExitUserError, "%s", err.Error())
		}
		if minRole, ok := requiredRoleForCommand(cmd); ok {
			if err := auth.RequireCLIRole(minRole); err != nil {
				return system.NewExitError(system.ExitUserError, "%s", err.Error())
			}
		}
		return nil
	},
}

// Execute runs the root command and exits on error.
func Execute() {
	rootCmd.Version = displayVersion()
	if err := rootCmd.Execute(); err != nil {
		ui.Fail("Error", err.Error())
		if code, ok := resolveExitCode(err); ok && code >= 0 {
			exitFunc(code)
			return
		}
		exitFunc(system.ExitUserError)
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

func displayVersion() string {
	if Commit == "" || Commit == "none" {
		return Version
	}
	return Version + " (" + Commit + ")"
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose debug output to stderr")
	cobra.OnInitialize(func() {
		system.InitLogger(verbose)
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:       "completion [bash|zsh|fish]",
		Short:     "Generate shell completion scripts",
		Hidden:    true,
		ValidArgs: []string{"bash", "zsh", "fish"},
		Args:      cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			default:
				return system.NewExitError(system.ExitUserError, "unsupported shell %s", args[0])
			}
		},
	})
}

func configureDefaultWorkspaceRoot(cmd *cobra.Command) {
	if cmd == nil {
		return
	}
	if _, ok := os.LookupEnv("GRADIENT_WORKSPACE_ROOT"); ok {
		return
	}
	if !shouldUseUserWorkspaceRoot(cmd) {
		return
	}
	_ = os.Setenv("GRADIENT_WORKSPACE_ROOT", workspaceUserRoot())
}

func shouldUseUserWorkspaceRoot(cmd *cobra.Command) bool {
	path := strings.TrimSpace(strings.TrimPrefix(cmd.CommandPath(), "concave"))
	switch path {
	case "serve":
		return false
	default:
		return true
	}
}

func isExemptCommand(cmd *cobra.Command) bool {
	if cmd == nil {
		return true
	}
	path := strings.TrimSpace(strings.TrimPrefix(cmd.CommandPath(), "concave"))
	switch path {
	case "", "check", "doctor", "completion", "whoami":
		return true
	default:
		return false
	}
}

func requiredRoleForCommand(cmd *cobra.Command) (auth.Role, bool) {
	switch strings.TrimSpace(strings.TrimPrefix(cmd.CommandPath(), "concave")) {
	case "status", "list", "logs", "changelog", "fleet status", "fleet peers", "node status", "env status", "env diff", "env baseline show", "resolver status", "resolver logs", "mesh status", "mesh logs", "gpu check", "gpu info":
		return auth.RoleViewer, true
	case "lab", "shell", "exec":
		return auth.RoleDeveloper, true
	case "workspace status":
		return auth.RoleViewer, true
	case "workspace backup", "workspace prune", "install", "remove", "start", "stop", "restart", "update", "rollback", "team create", "team delete", "team add-user", "team remove-user", "team list", "team status", "node set", "env export", "env apply", "env rollback", "env baseline set":
		return auth.RoleOperator, true
	case "serve", "gpu setup", "driver-wizard", "setup", "upgrade", "self-update", "resolver restart", "mesh restart":
		return auth.RoleAdmin, true
	default:
		return 0, false
	}
}
