package docker

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gradient-Linux/concave/internal/config"
	"github.com/Gradient-Linux/concave/internal/suite"
	"github.com/Gradient-Linux/concave/internal/workspace"
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

func TestWriteRawComposeRenameFailureRemovesTempFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := workspace.EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout() error = %v", err)
	}

	previous := runCombinedOutput
	runCombinedOutput = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return []byte("ok"), nil
	}
	defer func() { runCombinedOutput = previous }()

	path := workspace.ComposePath("boosting")
	if err := os.RemoveAll(path); err != nil {
		t.Fatalf("RemoveAll() error = %v", err)
	}
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	_, err := WriteRawCompose(context.Background(), "boosting", []byte("services:\n  demo:\n    image: busybox\n"))
	if err == nil {
		t.Fatal("expected rename error")
	}
	if _, statErr := os.Stat(path + ".tmp"); !os.IsNotExist(statErr) {
		t.Fatalf("expected temp file removal, got %v", statErr)
	}
}

func TestClientAndImageHelpers(t *testing.T) {
	previousCombined := runCombinedOutput
	previousStreaming := runStreaming
	t.Cleanup(func() {
		runCombinedOutput = previousCombined
		runStreaming = previousStreaming
	})

	var combinedCalls []string
	runCombinedOutput = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		combinedCalls = append(combinedCalls, name+" "+strings.Join(args, " "))
		switch args[0] {
		case "inspect":
			return []byte("running\n"), nil
		case "image":
			return []byte("present"), nil
		default:
			return []byte("ok"), nil
		}
	}
	runStreaming = func(ctx context.Context, name string, args []string, onLine func(string)) error {
		if onLine != nil {
			onLine("layer 1")
			onLine("layer 2")
		}
		return nil
	}

	if err := Run(context.Background(), "busybox", "echo", "ok"); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if err := Exec(context.Background(), "container", "echo", "ok"); err != nil {
		t.Fatalf("Exec() error = %v", err)
	}
	var lines []string
	if err := Pull(context.Background(), "busybox", func(line string) { lines = append(lines, line) }); err != nil {
		t.Fatalf("Pull() error = %v", err)
	}
	if err := ComposeUp(context.Background(), "/tmp/demo.yml", true); err != nil {
		t.Fatalf("ComposeUp() error = %v", err)
	}
	if err := ComposeDown(context.Background(), "/tmp/demo.yml"); err != nil {
		t.Fatalf("ComposeDown() error = %v", err)
	}
	status, err := ContainerStatus(context.Background(), "container")
	if err != nil || status != "running" {
		t.Fatalf("ContainerStatus() = %q, %v", status, err)
	}
	if err := PullWithProgress(context.Background(), "busybox:1.36", nil); err != nil {
		t.Fatalf("PullWithProgress() error = %v", err)
	}
	if err := RevertToPrevious("busybox:1.36"); err != nil {
		t.Fatalf("RevertToPrevious() error = %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("expected streamed lines, got %#v", lines)
	}
	if len(combinedCalls) == 0 {
		t.Fatal("expected docker commands to be invoked")
	}
}

func TestClientHelpersWrapErrors(t *testing.T) {
	previousCombined := runCombinedOutput
	previousStreaming := runStreaming
	t.Cleanup(func() {
		runCombinedOutput = previousCombined
		runStreaming = previousStreaming
	})

	runCombinedOutput = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return nil, errors.New("boom")
	}
	runStreaming = func(ctx context.Context, name string, args []string, onLine func(string)) error {
		return errors.New("boom")
	}

	tests := []struct {
		name string
		run  func() error
		want string
	}{
		{name: "run", run: func() error { return Run(context.Background(), "busybox", "echo", "ok") }, want: "docker run busybox"},
		{name: "exec", run: func() error { return Exec(context.Background(), "container", "echo", "ok") }, want: "docker exec container"},
		{name: "pull", run: func() error { return Pull(context.Background(), "busybox", nil) }, want: "docker pull busybox"},
		{name: "compose up", run: func() error { return ComposeUp(context.Background(), "/tmp/demo.yml", true) }, want: "docker compose up /tmp/demo.yml"},
		{name: "compose down", run: func() error { return ComposeDown(context.Background(), "/tmp/demo.yml") }, want: "docker compose down /tmp/demo.yml"},
		{name: "status", run: func() error { _, err := ContainerStatus(context.Background(), "container"); return err }, want: "docker inspect container"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run()
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error %q does not contain %q", err, tt.want)
			}
		})
	}
}

func TestImageHelpersHandleMissingAndFailedTags(t *testing.T) {
	previousCombined := runCombinedOutput
	t.Cleanup(func() { runCombinedOutput = previousCombined })

	calls := make([]string, 0, 4)
	runCombinedOutput = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		call := name + " " + strings.Join(args, " ")
		calls = append(calls, call)
		switch {
		case strings.Contains(call, "image inspect busybox:1.36"):
			return nil, errors.New("missing")
		case strings.Contains(call, "image inspect busybox:gradient-previous"):
			return nil, errors.New("missing previous")
		default:
			return []byte("ok"), nil
		}
	}

	if err := TagAsPrevious("busybox:1.36"); err != nil {
		t.Fatalf("TagAsPrevious() error = %v", err)
	}
	if err := RevertToPrevious("busybox:1.36"); err == nil {
		t.Fatal("expected RevertToPrevious() error")
	}
	if len(calls) == 0 {
		t.Fatal("expected docker image calls")
	}
}

func TestImageHelpersWrapTagFailures(t *testing.T) {
	previousCombined := runCombinedOutput
	t.Cleanup(func() { runCombinedOutput = previousCombined })

	runCombinedOutput = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		call := name + " " + strings.Join(args, " ")
		switch {
		case strings.Contains(call, "image inspect"):
			return []byte("present"), nil
		case strings.Contains(call, " tag "):
			return nil, errors.New("tag failed")
		default:
			return nil, errors.New("unexpected")
		}
	}

	if err := TagAsPrevious("busybox:1.36"); err == nil || !strings.Contains(err.Error(), "docker tag busybox:1.36 busybox:gradient-previous") {
		t.Fatalf("TagAsPrevious() error = %v", err)
	}
	if err := RevertToPrevious("busybox:1.36"); err == nil || !strings.Contains(err.Error(), "docker tag busybox:gradient-previous busybox:1.36") {
		t.Fatalf("RevertToPrevious() error = %v", err)
	}
}

func TestValidateComposeAndWriteSuiteCompose(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := workspace.EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout() error = %v", err)
	}

	previous := runCombinedOutput
	runCombinedOutput = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return []byte("ok"), nil
	}
	defer func() { runCombinedOutput = previous }()

	if err := ValidateCompose(context.Background(), "/tmp/demo.yml"); err != nil {
		t.Fatalf("ValidateCompose() error = %v", err)
	}

	s, err := suite.Get("boosting")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	path, err := WriteSuiteCompose(context.Background(), s)
	if err != nil {
		t.Fatalf("WriteSuiteCompose() error = %v", err)
	}
	if path != workspace.ComposePath("boosting") {
		t.Fatalf("WriteSuiteCompose() path = %q", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected compose file at %s: %v", path, err)
	}
}

func TestReadTemplateUsesLookupCandidates(t *testing.T) {
	previousRead := readTemplateFile
	t.Cleanup(func() { readTemplateFile = previousRead })

	var attempted []string
	readTemplateFile = func(path string) ([]byte, error) {
		attempted = append(attempted, filepath.Clean(path))
		if strings.Contains(path, "boosting.compose.yml") {
			return []byte("services: {}\n"), nil
		}
		return nil, os.ErrNotExist
	}

	data, err := readTemplate("boosting")
	if err != nil {
		t.Fatalf("readTemplate() error = %v", err)
	}
	if string(data) != "services: {}\n" {
		t.Fatalf("readTemplate() data = %q", string(data))
	}
	if len(attempted) == 0 {
		t.Fatal("expected template lookup attempts")
	}
}

func TestRenderSuiteComposeMissingTemplate(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := workspace.EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout() error = %v", err)
	}

	previousRead := readTemplateFile
	t.Cleanup(func() { readTemplateFile = previousRead })
	readTemplateFile = func(path string) ([]byte, error) {
		return nil, os.ErrNotExist
	}

	_, err := RenderSuiteCompose(suite.Suite{Name: "missing", ComposeTemplate: "missing"})
	if err == nil {
		t.Fatal("expected template error")
	}
	if !strings.Contains(err.Error(), "missing.compose.yml") {
		t.Fatalf("expected missing template error, got %v", err)
	}
}
