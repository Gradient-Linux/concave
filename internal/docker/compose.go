package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gradient-linux/concave/internal/config"
	"github.com/gradient-linux/concave/internal/suite"
	"github.com/gradient-linux/concave/internal/workspace"
)

const composeNetwork = "gradient-network"

// RenderSuiteCompose renders a suite template with workspace and image substitutions.
func RenderSuiteCompose(s suite.Suite) ([]byte, error) {
	path, err := templatePath(s.ComposeTemplate)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	rendered := strings.ReplaceAll(string(data), "{{WORKSPACE_ROOT}}", workspace.Root())
	rendered = strings.ReplaceAll(rendered, "{{COMPOSE_NETWORK}}", composeNetwork)

	versions, err := config.LoadVersions()
	if err != nil {
		return nil, err
	}
	for _, container := range s.Containers {
		if version, ok := config.GetImageVersion(versions, s.Name, container.Name); ok && version.Current != "" {
			rendered = strings.ReplaceAll(rendered, "image: "+container.Image, "image: "+version.Current)
		}
	}

	return []byte(rendered), nil
}

// ValidateCompose validates a rendered compose file path with docker compose config --quiet.
func ValidateCompose(ctx context.Context, path string) error {
	if _, err := runCombinedOutput(ctx, "docker", "compose", "-f", path, "config", "--quiet"); err != nil {
		return fmt.Errorf("docker compose config %s: %w", path, err)
	}
	return nil
}

// WriteSuiteCompose renders, validates, and installs a suite compose file.
func WriteSuiteCompose(ctx context.Context, s suite.Suite) (string, error) {
	data, err := RenderSuiteCompose(s)
	if err != nil {
		return "", err
	}
	return WriteRawCompose(ctx, s.Name, data)
}

// WriteRawCompose writes arbitrary compose content to the workspace after validation.
func WriteRawCompose(ctx context.Context, name string, data []byte) (string, error) {
	if err := workspace.EnsureLayout(); err != nil {
		return "", err
	}

	path := workspace.ComposePath(name)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", tmp, err)
	}

	validateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ValidateCompose(validateCtx, tmp); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return "", fmt.Errorf("rename %s to %s: %w", tmp, path, err)
	}

	return path, nil
}

func templatePath(name string) (string, error) {
	if _, filename, _, ok := runtime.Caller(0); ok {
		root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
		path := filepath.Join(root, "templates", name+".compose.yml")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	path := filepath.Join("templates", name+".compose.yml")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("template %s.compose.yml not found", name)
}
