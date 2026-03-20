package config

import (
	"testing"

	"github.com/gradientlinux/concave/internal/workspace"
)

func TestStateRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := workspace.EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout() error = %v", err)
	}

	if err := AddInstalled("boosting"); err != nil {
		t.Fatalf("AddInstalled() error = %v", err)
	}
	if err := AddInstalled("flow"); err != nil {
		t.Fatalf("AddInstalled() error = %v", err)
	}

	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}
	if len(state.Installed) != 2 {
		t.Fatalf("expected 2 installed suites, got %d", len(state.Installed))
	}

	if err := RemoveInstalled("boosting"); err != nil {
		t.Fatalf("RemoveInstalled() error = %v", err)
	}
	ok, err := IsInstalled("boosting")
	if err != nil {
		t.Fatalf("IsInstalled() error = %v", err)
	}
	if ok {
		t.Fatal("expected boosting to be removed")
	}
}

func TestVersionsRoundTripAndSwap(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := workspace.EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout() error = %v", err)
	}

	versions := Versions{}
	SetImageVersion(versions, "boosting", "gradient-boost-core", "python:3.12-slim", "")
	SetImageVersion(versions, "boosting", "gradient-boost-lab", "quay.io/jupyter/base-notebook:python-3.11.6", "quay.io/jupyter/base-notebook:python-3.10.13")
	if err := SaveVersions(versions); err != nil {
		t.Fatalf("SaveVersions() error = %v", err)
	}

	loaded, err := LoadVersions()
	if err != nil {
		t.Fatalf("LoadVersions() error = %v", err)
	}
	if _, ok := GetImageVersion(loaded, "boosting", "gradient-boost-core"); !ok {
		t.Fatal("expected version entry for gradient-boost-core")
	}
	if err := SwapPrevious(loaded, "boosting"); err == nil {
		t.Fatal("expected swap error when previous tag is missing")
	}
}
