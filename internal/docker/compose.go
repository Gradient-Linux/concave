package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Gradient-Linux/concave/internal/config"
	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/Gradient-Linux/concave/internal/workspace"
)

const composeNetwork = "gradient-network"

var readTemplateFile = os.ReadFile

// RenderSuiteCompose renders a suite template with workspace and image substitutions.
func RenderSuiteCompose(s suite.Suite) ([]byte, error) {
	data, err := readTemplate(s.ComposeTemplate)
	if err != nil {
		return nil, err
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

func readTemplate(name string) ([]byte, error) {
	filename := name + ".compose.yml"
	candidates := []string{filepath.Join("templates", filename)}

	if _, sourceFile, _, ok := runtime.Caller(0); ok {
		repoRoot := filepath.Clean(filepath.Join(filepath.Dir(sourceFile), "..", ".."))
		candidates = append(candidates, filepath.Join(repoRoot, "templates", filename))
	}

	if executable, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(executable), "templates", filename))
	}

	var failures []string
	for _, candidate := range uniqueCandidates(candidates) {
		data, err := readTemplateFile(candidate)
		if err == nil {
			return data, nil
		}
		failures = append(failures, fmt.Sprintf("%s: %v", candidate, err))
	}

	return nil, fmt.Errorf("read template %s: %s", filename, strings.Join(failures, "; "))
}

func uniqueCandidates(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		result = append(result, path)
	}
	return result
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
