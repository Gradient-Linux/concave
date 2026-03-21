package suite

import (
	"os"
	"strings"
	"testing"
)

func TestRegistryContainsAllSuites(t *testing.T) {
	names := Names()
	expected := map[string]bool{
		"boosting": false,
		"flow":     false,
		"forge":    false,
		"neural":   false,
	}
	for _, name := range names {
		if _, ok := expected[name]; ok {
			expected[name] = true
		}
	}
	for name, seen := range expected {
		if !seen {
			t.Fatalf("missing suite %s", name)
		}
	}
}

func TestBuildInstallPlan(t *testing.T) {
	plan, err := BuildInstallPlan("boosting")
	if err != nil {
		t.Fatalf("BuildInstallPlan() error = %v", err)
	}
	if plan.PrimaryContainer != "gradient-boost-core" {
		t.Fatalf("unexpected primary container %q", plan.PrimaryContainer)
	}
	if plan.JupyterContainer != "gradient-boost-lab" {
		t.Fatalf("unexpected jupyter container %q", plan.JupyterContainer)
	}
	if len(plan.Images) != 3 {
		t.Fatalf("unexpected image count %d", len(plan.Images))
	}
}

func TestAllHelpersAndForgeCompose(t *testing.T) {
	suites := All()
	if len(suites) != 4 {
		t.Fatalf("All() returned %d suites", len(suites))
	}

	boosting, err := Get("boosting")
	if err != nil {
		t.Fatalf("Get(boosting) error = %v", err)
	}
	names := ContainerNames(boosting)
	if len(names) != 3 || names[0] != "gradient-boost-core" {
		t.Fatalf("ContainerNames(boosting) = %#v", names)
	}
	if ports := SuitePorts(boosting); ports != "5000, 8888" {
		t.Fatalf("SuitePorts(boosting) = %q", ports)
	}
	if _, err := Get("missing"); err == nil {
		t.Fatal("expected missing suite error")
	}

	compose, err := BuildForgeCompose([]string{"gradient-boost-core", "gradient-flow-mlflow"})
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
}

func TestSelectForgeComponents(t *testing.T) {
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error = %v", err)
	}
	if _, err := w.WriteString("1,4\n"); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	_ = w.Close()
	os.Stdin = r
	defer func() {
		_ = r.Close()
		os.Stdin = oldStdin
	}()

	selected := SelectForgeComponents()
	if len(selected) != 2 || selected[0] != "gradient-boost-core" || selected[1] != "gradient-neural-torch" {
		t.Fatalf("SelectForgeComponents() = %#v", selected)
	}
}
