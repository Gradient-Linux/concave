package suite

import (
	"strings"
	"testing"
)

func TestRegistryContainsAllSuites(t *testing.T) {
	for _, name := range []string{"boosting", "neural", "flow", "forge"} {
		s, err := Get(name)
		if err != nil {
			t.Fatalf("Get(%s) error = %v", name, err)
		}
		if len(s.Containers) == 0 || len(s.Ports) == 0 || len(s.Volumes) == 0 {
			t.Fatalf("suite %s is incomplete: %#v", name, s)
		}
	}
}

func TestGetUnknownSuiteIncludesValidNames(t *testing.T) {
	_, err := Get("unknown")
	if err == nil {
		t.Fatal("expected error")
	}
	want := "unknown suite: unknown. Valid suites: boosting, neural, flow, forge"
	if err.Error() != want {
		t.Fatalf("Get() error = %q, want %q", err.Error(), want)
	}
}

func TestPickComponentsDeduplicatesJupyter(t *testing.T) {
	oldPrompt := promptChecklist
	t.Cleanup(func() { promptChecklist = oldPrompt })

	promptChecklist = func(items []string) []string {
		return []string{
			"Boosting | JupyterLab (~1 GB, shared with Neural)",
			"Neural | JupyterLab (~1 GB, shared with Boosting)",
			"Flow | Model serving (~800 MB)",
		}
	}

	selection, err := PickComponents()
	if err != nil {
		t.Fatalf("PickComponents() error = %v", err)
	}
	count := 0
	for _, container := range selection.Containers {
		if strings.Contains(container.Name, "lab") {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected a single deduplicated Jupyter container, got %#v", selection.Containers)
	}
}

func TestPickComponentsRejectsEmptySelection(t *testing.T) {
	oldPrompt := promptChecklist
	t.Cleanup(func() { promptChecklist = oldPrompt })

	promptChecklist = func(items []string) []string { return nil }
	if _, err := PickComponents(); err == nil || err.Error() != "no components selected" {
		t.Fatalf("PickComponents() error = %v", err)
	}
}

func TestBuildForgeComposeFiltersSelectedServices(t *testing.T) {
	selection, err := SelectionFromContainerNames(
		[]string{"gradient-boost-core", "gradient-flow-mlflow"},
		map[string]string{"gradient-flow-mlflow": "ghcr.io/mlflow/mlflow:2.15"},
	)
	if err != nil {
		t.Fatalf("SelectionFromContainerNames() error = %v", err)
	}

	compose, err := BuildForgeCompose(selection)
	if err != nil {
		t.Fatalf("BuildForgeCompose() error = %v", err)
	}
	text := string(compose)
	if !strings.Contains(text, "gradient-boost-core:") || !strings.Contains(text, "gradient-flow-mlflow:") {
		t.Fatalf("forge compose missing selected services:\n%s", text)
	}
	if strings.Contains(text, "profiles: [\"disabled\"]") {
		t.Fatalf("forge compose still contains disabled profiles:\n%s", text)
	}
	if strings.Contains(text, "gradient-neural-torch:") {
		t.Fatalf("forge compose should exclude unselected services:\n%s", text)
	}
	if !strings.Contains(text, "ghcr.io/mlflow/mlflow:2.15") {
		t.Fatalf("expected image override in compose:\n%s", text)
	}
}
