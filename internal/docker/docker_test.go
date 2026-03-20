package docker

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/gradientlinux/concave/internal/config"
	"github.com/gradientlinux/concave/internal/suite"
	"github.com/gradientlinux/concave/internal/workspace"
)

func TestPreviousImageTag(t *testing.T) {
	if got := previousImageTag("python:3.12-slim"); got != "python:gradient-previous" {
		t.Fatalf("previousImageTag() = %q", got)
	}
	if got := previousImageTag("ghcr.io/mlflow/mlflow:2.14"); got != "ghcr.io/mlflow/mlflow:gradient-previous" {
		t.Fatalf("previousImageTag() = %q", got)
	}
}

func TestRenderSuiteComposeUsesStoredVersions(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := workspace.EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout() error = %v", err)
	}

	versions := config.Versions{}
	config.SetImageVersion(versions, "boosting", "gradient-boost-core", "python:3.12-alpine", "")
	if err := config.SaveVersions(versions); err != nil {
		t.Fatalf("SaveVersions() error = %v", err)
	}

	s, err := suite.Get("boosting")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	rendered, err := RenderSuiteCompose(s)
	if err != nil {
		t.Fatalf("RenderSuiteCompose() error = %v", err)
	}
	text := string(rendered)
	if !strings.Contains(text, "python:3.12-alpine") {
		t.Fatalf("expected overridden image in compose:\n%s", text)
	}
	if !strings.Contains(text, workspace.Root()) {
		t.Fatalf("expected workspace root in compose:\n%s", text)
	}
}

func TestWriteRawComposeValidationFailureRemovesTempFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := workspace.EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout() error = %v", err)
	}

	previous := runCombinedOutput
	runCombinedOutput = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return nil, errors.New("invalid compose")
	}
	defer func() { runCombinedOutput = previous }()

	_, err := WriteRawCompose(context.Background(), "boosting", []byte("services:\n  demo:\n    image: busybox\n"))
	if err == nil {
		t.Fatal("expected validation error")
	}
	if _, statErr := os.Stat(workspace.ComposePath("boosting") + ".tmp"); !os.IsNotExist(statErr) {
		t.Fatalf("expected temp file removal, got %v", statErr)
	}
}
