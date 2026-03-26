package system

import (
	"context"
	"fmt"
	"os/exec"
)

// CommandAvailable reports whether an executable is available in PATH.
func CommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// DockerComposeAvailable reports whether `docker compose` is available.
func DockerComposeAvailable() (bool, error) {
	if !CommandAvailable("docker") {
		return false, nil
	}
	if _, err := runner.Run("docker", "compose", "version"); err != nil {
		return false, fmt.Errorf("docker compose version: %w", err)
	}
	return true, nil
}

// RestartOrEnableService enables and starts a systemd service.
func RestartOrEnableService(ctx context.Context, service string) error {
	return RunPrivileged(ctx, "Enable "+service+" service", "systemctl", "enable", "--now", service)
}
