package workspace

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureLayoutCreatesWorkspace(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout() error = %v", err)
	}

	for _, name := range []string{"data", "notebooks", "models", "outputs", "mlruns", "dags", "compose", "config", "backups"} {
		if _, err := os.Stat(filepath.Join(Root(), name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}

func TestStatusAndCleanOutputs(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout() error = %v", err)
	}

	if err := os.WriteFile(filepath.Join(Root(), "outputs", "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	usages, err := Status()
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if len(usages) == 0 {
		t.Fatal("expected non-empty usage slice")
	}

	if err := CleanOutputs(); err != nil {
		t.Fatalf("CleanOutputs() error = %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(Root(), "outputs"))
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty outputs dir, got %d entries", len(entries))
	}
}

func TestBackupCreatesArchive(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := EnsureLayout(); err != nil {
		t.Fatalf("EnsureLayout() error = %v", err)
	}

	if err := os.WriteFile(filepath.Join(Root(), "models", "model.bin"), []byte("model"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(Root(), "notebooks", "note.ipynb"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	archivePath, err := Backup()
	if err != nil {
		t.Fatalf("Backup() error = %v", err)
	}

	file, err := os.Open(archivePath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		t.Fatalf("NewReader() error = %v", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	found := map[string]bool{}
	for {
		header, err := tr.Next()
		if err != nil {
			break
		}
		found[header.Name] = true
	}
	if !found["models/model.bin"] || !found["notebooks/note.ipynb"] {
		t.Fatalf("archive missing expected files: %#v", found)
	}
}
