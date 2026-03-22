package system

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Gradient-Linux/concave/internal/ui"
)

var runPrivilegedCommand = func(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, "sudo", append([]string{name}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RunPrivileged runs a command with sudo and always prints what will run first.
func RunPrivileged(ctx context.Context, description string, name string, args ...string) error {
	ui.Info("Sudo", description)
	ui.Info("Running", "sudo "+name+" "+strings.Join(args, " "))
	if err := runPrivilegedCommand(ctx, name, args...); err != nil {
		return fmt.Errorf("%s failed: %w\nIf sudo is not configured, run manually:\n  sudo %s %s", description, err, name, strings.Join(args, " "))
	}
	return nil
}
