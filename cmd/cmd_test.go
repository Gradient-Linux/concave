package cmd

import (
	"testing"

	"github.com/gradientlinux/concave/internal/config"
	"github.com/gradientlinux/concave/internal/workspace"
)

func TestExtractLabURL(t *testing.T) {
	raw := "Currently running servers:\nhttp://localhost:8888/?token=abcdef :: /notebooks\n"
	got, err := extractLabURL(raw)
	if err != nil {
		t.Fatalf("extractLabURL() error = %v", err)
	}
	want := "http://127.0.0.1:8888/lab?token=abcdef"
	if got != want {
		t.Fatalf("extractLabURL() = %q, want %q", got, want)
	}
}

func TestTargetSuitesUsesInstalledState(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := workspace.EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout() error = %v", err)
	}
	if err := config.AddInstalled("boosting"); err != nil {
		t.Fatalf("AddInstalled() error = %v", err)
	}
	if err := config.AddInstalled("flow"); err != nil {
		t.Fatalf("AddInstalled() error = %v", err)
	}

	names, err := targetSuites(nil)
	if err != nil {
		t.Fatalf("targetSuites() error = %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 suites, got %d", len(names))
	}
}
