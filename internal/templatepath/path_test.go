package templatepath

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestCandidatesIncludeSharedInstallPaths(t *testing.T) {
	oldExecutablePath := executablePath
	t.Cleanup(func() { executablePath = oldExecutablePath })

	executablePath = func() (string, error) {
		return "/usr/local/bin/concave", nil
	}

	got := Candidates("boosting.compose.yml", "/work/concave/internal/docker/compose.go")
	want := []string{
		filepath.Join("templates", "boosting.compose.yml"),
		"/work/concave/templates/boosting.compose.yml",
		"/usr/local/bin/templates/boosting.compose.yml",
		"/usr/local/share/concave/templates/boosting.compose.yml",
		"/usr/share/concave/templates/boosting.compose.yml",
	}

	if len(got) != len(want) {
		t.Fatalf("Candidates() len = %d, want %d (%#v)", len(got), len(want), got)
	}
	for idx := range want {
		if got[idx] != want[idx] {
			t.Fatalf("Candidates()[%d] = %q, want %q", idx, got[idx], want[idx])
		}
	}
}

func TestCandidatesWorkWithoutCallerFile(t *testing.T) {
	oldExecutablePath := executablePath
	t.Cleanup(func() { executablePath = oldExecutablePath })

	executablePath = func() (string, error) {
		return "/opt/gradient/bin/concave", nil
	}

	got := Candidates("forge.compose.yml", "")
	if got[0] != filepath.Join("templates", "forge.compose.yml") {
		t.Fatalf("Candidates()[0] = %q", got[0])
	}
	if got[1] != "/opt/gradient/bin/templates/forge.compose.yml" {
		t.Fatalf("Candidates()[1] = %q", got[1])
	}
	if got[2] != "/opt/gradient/share/concave/templates/forge.compose.yml" {
		t.Fatalf("Candidates()[2] = %q", got[2])
	}
}

func TestCandidatesHandleMissingExecutable(t *testing.T) {
	oldExecutablePath := executablePath
	t.Cleanup(func() { executablePath = oldExecutablePath })

	executablePath = func() (string, error) {
		return "", errors.New("no executable")
	}

	got := Candidates("flow.compose.yml", "/repo/internal/docker/compose.go")
	if len(got) != 4 {
		t.Fatalf("Candidates() len = %d, want 4 (%#v)", len(got), got)
	}
	if got[1] != "/repo/templates/flow.compose.yml" {
		t.Fatalf("Candidates()[1] = %q", got[1])
	}
}
