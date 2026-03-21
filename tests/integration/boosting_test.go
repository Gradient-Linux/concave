package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBoostingLifecycle(t *testing.T) {
	requireIntegration(t)

	repoRoot := resolveRepoRoot(t)
	binaryPath := buildConcaveBinary(t, repoRoot)
	homeDir := t.TempDir()
	binDir := t.TempDir()
	openLog := filepath.Join(t.TempDir(), "opened-url.log")
	writeScript(t, filepath.Join(binDir, "xdg-open"), "#!/bin/sh\nprintf '%s\\n' \"$1\" >> \"$OPENED_URL_LOG\"\n")

	env := append(os.Environ(),
		"HOME="+homeDir,
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"OPENED_URL_LOG="+openLog,
	)

	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cleanupCancel()
	defer func() {
		_, _ = runConcave(cleanupCtx, repoRoot, binaryPath, env, "", "stop", "boosting")
		_, _ = runConcave(cleanupCtx, repoRoot, binaryPath, env, "y\n", "remove", "boosting")
	}()

	runCtx, cancel := context.WithTimeout(context.Background(), 25*time.Minute)
	defer cancel()

	if out, err := runConcave(runCtx, repoRoot, binaryPath, env, "", "workspace", "init"); err != nil {
		t.Fatalf("workspace init failed: %v\n%s", err, out)
	}
	if out, err := runConcave(runCtx, repoRoot, binaryPath, env, "", "install", "boosting"); err != nil {
		t.Fatalf("install boosting failed: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(homeDir, "gradient", "compose", "boosting.compose.yml")); err != nil {
		t.Fatalf("expected rendered compose file: %v", err)
	}
	if out, err := runConcave(runCtx, repoRoot, binaryPath, env, "", "start", "boosting"); err != nil {
		t.Fatalf("start boosting failed: %v\n%s", err, out)
	}

	waitForLab(t, repoRoot, binaryPath, env, openLog)
}

func requireIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("CONCAVE_INTEGRATION") != "1" {
		t.Skip("set CONCAVE_INTEGRATION=1 to run Docker-backed integration tests")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skipf("docker not found: %v", err)
	}
	cmd := exec.Command("docker", "info")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("docker info failed: %v\n%s", err, string(out))
	}
}

func resolveRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	root := filepath.Clean(filepath.Join(wd, "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repo root lookup failed: %v", err)
	}
	return root
}

func buildConcaveBinary(t *testing.T, repoRoot string) string {
	t.Helper()
	binaryPath := filepath.Join(t.TempDir(), "concave")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(out))
	}
	return binaryPath
}

func writeScript(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}

func runConcave(ctx context.Context, repoRoot, binaryPath string, env []string, stdin string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	cmd.Dir = repoRoot
	cmd.Env = env
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func waitForLab(t *testing.T, repoRoot, binaryPath string, env []string, openLog string) {
	t.Helper()

	deadline := time.Now().Add(3 * time.Minute)
	var lastOutput string
	for time.Now().Before(deadline) {
		_ = os.WriteFile(openLog, nil, 0o644)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		out, err := runConcave(ctx, repoRoot, binaryPath, env, "", "lab", "--suite", "boosting")
		cancel()
		lastOutput = out
		if err == nil {
			data, readErr := os.ReadFile(openLog)
			if readErr != nil {
				t.Fatalf("ReadFile(%s) error = %v", openLog, readErr)
			}
			url := strings.TrimSpace(string(data))
			if strings.Contains(url, "/lab?token=") {
				return
			}
			t.Fatalf("lab opened unexpected URL %q", url)
		}
		time.Sleep(5 * time.Second)
	}

	t.Fatalf("lab never became ready; last output:\n%s", lastOutput)
}
