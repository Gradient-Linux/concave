package suite

import "testing"

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
