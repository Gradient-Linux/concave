package docker

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

var (
	runCombinedOutput = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return exec.CommandContext(ctx, name, args...).CombinedOutput()
	}
	runStreaming = func(ctx context.Context, name string, args []string, onLine func(string)) error {
		cmd := exec.CommandContext(ctx, name, args...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("stdout pipe: %w", err)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("stderr pipe: %w", err)
		}
		if err := cmd.Start(); err != nil {
			return err
		}

		for _, stream := range []io.Reader{stdout, stderr} {
			scanner := bufio.NewScanner(stream)
			for scanner.Scan() {
				if onLine != nil {
					onLine(scanner.Text())
				}
			}
		}

		if err := cmd.Wait(); err != nil {
			return err
		}
		return nil
	}
)

// Run executes docker run --rm with the supplied image and arguments.
func Run(ctx context.Context, image string, args ...string) error {
	command := append([]string{"run", "--rm", image}, args...)
	if _, err := runCombinedOutput(ctx, "docker", command...); err != nil {
		return fmt.Errorf("docker run %s: %w", image, err)
	}
	return nil
}

// Exec executes a command inside a running container.
func Exec(ctx context.Context, container string, cmd ...string) error {
	command := append([]string{"exec", container}, cmd...)
	if _, err := runCombinedOutput(ctx, "docker", command...); err != nil {
		return fmt.Errorf("docker exec %s: %w", container, err)
	}
	return nil
}

// Pull pulls an image and streams progress lines to the callback.
func Pull(ctx context.Context, image string, onProgress func(line string)) error {
	if err := runStreaming(ctx, "docker", []string{"pull", image}, onProgress); err != nil {
		return fmt.Errorf("docker pull %s: %w", image, err)
	}
	return nil
}

// ComposeUp starts a compose application.
func ComposeUp(ctx context.Context, composePath string, detach bool) error {
	command := []string{"compose", "-f", composePath, "up"}
	if detach {
		command = append(command, "-d")
	}
	if _, err := runCombinedOutput(ctx, "docker", command...); err != nil {
		return fmt.Errorf("docker compose up %s: %w", composePath, err)
	}
	return nil
}

// ComposeDown stops and removes a compose application.
func ComposeDown(ctx context.Context, composePath string) error {
	command := []string{"compose", "-f", composePath, "down"}
	if _, err := runCombinedOutput(ctx, "docker", command...); err != nil {
		return fmt.Errorf("docker compose down %s: %w", composePath, err)
	}
	return nil
}

// ContainerStatus returns a container's current status.
func ContainerStatus(ctx context.Context, name string) (string, error) {
	out, err := runCombinedOutput(ctx, "docker", "inspect", "-f", "{{.State.Status}}", name)
	if err != nil {
		return "", fmt.Errorf("docker inspect %s: %w", name, err)
	}
	return strings.TrimSpace(string(out)), nil
}
