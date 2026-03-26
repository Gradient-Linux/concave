package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Gradient-Linux/concave/internal/ui"
)

func ensureDockerRuntime(ctx context.Context, action string) error {
	installedOrUpdated := false

	if !systemCommandAvailable("docker") {
		ui.Warn("Docker", "docker CLI not found in PATH")
		if !uiConfirm("Docker is required for " + action + ". Install Docker now with sudo?") {
			return wrapDockerError(fmt.Errorf("docker is required for %s; install Docker and retry", action))
		}
		if err := installDockerPackages(ctx); err != nil {
			return wrapDockerError(err)
		}
		installedOrUpdated = true
	}

	composeOK, err := systemDockerCompose()
	if err != nil {
		ui.Warn("Docker Compose", err.Error())
	}
	if !composeOK {
		if !uiConfirm("Docker Compose v2 is required for " + action + ". Install or update it now with sudo?") {
			return wrapDockerError(fmt.Errorf("docker compose v2 is required for %s; install it and retry", action))
		}
		if err := installDockerPackages(ctx); err != nil {
			return wrapDockerError(err)
		}
		installedOrUpdated = true
	}

	running, err := systemDockerRunning()
	if err != nil {
		ui.Warn("Docker", err.Error())
	}
	if !running {
		if !uiConfirm("Docker is not running. Start and enable the docker service now with sudo?") {
			return wrapDockerError(fmt.Errorf("docker service is required for %s; start docker and retry", action))
		}
		if err := systemRunPrivileged(ctx, "Enable docker service", "systemctl", "enable", "--now", "docker"); err != nil {
			return wrapDockerError(err)
		}
	}

	inGroup, err := systemUserInDockerGroup()
	if err != nil {
		ui.Warn("Docker group", err.Error())
	}
	if !inGroup {
		user := currentUsername()
		if !uiConfirm("Your user is not in the docker group. Add " + user + " now with sudo?") {
			return wrapUserError(fmt.Errorf("user %s is not in the docker group; add it and retry", user))
		}
		if err := systemRunPrivileged(ctx, "Add "+user+" to docker group", "usermod", "-aG", "docker", user); err != nil {
			return wrapUserError(err)
		}
		return wrapUserError(fmt.Errorf("added %s to the docker group; log out and back in, then retry concave %s", user, strings.TrimSpace(action)))
	}

	if installedOrUpdated {
		running, err = systemDockerRunning()
		if err != nil || !running {
			if err == nil {
				err = fmt.Errorf("docker did not become available after installation")
			}
			return wrapDockerError(err)
		}
	}

	ui.Pass("Docker", "runtime ready")
	return nil
}

func installDockerPackages(ctx context.Context) error {
	if err := systemRunPrivileged(ctx, "Refresh apt package index", "apt-get", "update"); err != nil {
		return err
	}
	if err := systemRunPrivileged(ctx, "Install Docker Engine", "env", "DEBIAN_FRONTEND=noninteractive", "apt-get", "install", "-y", "docker.io", "docker-compose-v2"); err != nil {
		ui.Warn("Docker", "docker-compose-v2 install failed, retrying with docker-compose-plugin")
		if fallbackErr := systemRunPrivileged(ctx, "Install Docker Engine", "env", "DEBIAN_FRONTEND=noninteractive", "apt-get", "install", "-y", "docker.io", "docker-compose-plugin"); fallbackErr != nil {
			return fmt.Errorf("install docker packages: primary=%v fallback=%w", err, fallbackErr)
		}
	}
	if err := systemRunPrivileged(ctx, "Enable docker service", "systemctl", "enable", "--now", "docker"); err != nil {
		return err
	}
	return nil
}
