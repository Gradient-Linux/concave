package suite

import (
	"context"
	"errors"
	"testing"

	"github.com/Gradient-Linux/concave/internal/config"
)

func TestInstallRunsHappyPath(t *testing.T) {
	oldLoadState := loadInstalledState
	oldInstalled := suiteInstalled
	oldLoadManifest := loadVersionManifest
	oldSaveManifest := saveVersionManifest
	oldAddSuite := addInstalledSuite
	oldPull := pullImageWithRollback
	oldWriteCompose := writeComposeFile
	oldWriteRaw := writeRawComposeFile
	oldConflicts := checkConflicts
	t.Cleanup(func() {
		loadInstalledState = oldLoadState
		suiteInstalled = oldInstalled
		loadVersionManifest = oldLoadManifest
		saveVersionManifest = oldSaveManifest
		addInstalledSuite = oldAddSuite
		pullImageWithRollback = oldPull
		writeComposeFile = oldWriteCompose
		writeRawComposeFile = oldWriteRaw
		checkConflicts = oldConflicts
	})

	loadInstalledState = func() (config.State, error) { return config.State{}, nil }
	suiteInstalled = func(name string) (bool, error) { return false, nil }
	loadVersionManifest = func() (config.VersionManifest, error) { return config.VersionManifest{}, nil }
	saveVersionManifest = func(manifest config.VersionManifest) error { return nil }
	addInstalledSuite = func(name string) error { return nil }
	pullImageWithRollback = func(ctx context.Context, image string, onProgress func(string)) error { return nil }
	writeComposeFile = func(name string) (string, error) { return "/tmp/" + name + ".compose.yml", nil }
	checkConflicts = func(s Suite, installed []string) ([]PortConflict, error) { return []PortConflict{}, nil }

	if err := Install(context.Background(), "boosting", InstallOptions{}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
}

func TestInstallStopsOnConflict(t *testing.T) {
	oldLoadState := loadInstalledState
	oldInstalled := suiteInstalled
	oldConflicts := checkConflicts
	oldPull := pullImageWithRollback
	t.Cleanup(func() {
		loadInstalledState = oldLoadState
		suiteInstalled = oldInstalled
		checkConflicts = oldConflicts
		pullImageWithRollback = oldPull
	})

	loadInstalledState = func() (config.State, error) { return config.State{Installed: []string{"flow"}}, nil }
	suiteInstalled = func(name string) (bool, error) { return false, nil }
	checkConflicts = func(s Suite, installed []string) ([]PortConflict, error) {
		return []PortConflict{{Port: 8080, ExistingSuite: "flow", Service: "Airflow"}}, nil
	}
	pullImageWithRollback = func(ctx context.Context, image string, onProgress func(string)) error {
		t.Fatal("pull should not run when conflicts exist")
		return nil
	}

	if err := Install(context.Background(), "neural", InstallOptions{}); err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestInstallStopsBeforeComposeOnPullFailure(t *testing.T) {
	oldLoadState := loadInstalledState
	oldInstalled := suiteInstalled
	oldLoadManifest := loadVersionManifest
	oldSaveManifest := saveVersionManifest
	oldAddSuite := addInstalledSuite
	oldPull := pullImageWithRollback
	oldWriteCompose := writeComposeFile
	oldConflicts := checkConflicts
	t.Cleanup(func() {
		loadInstalledState = oldLoadState
		suiteInstalled = oldInstalled
		loadVersionManifest = oldLoadManifest
		saveVersionManifest = oldSaveManifest
		addInstalledSuite = oldAddSuite
		pullImageWithRollback = oldPull
		writeComposeFile = oldWriteCompose
		checkConflicts = oldConflicts
	})

	loadInstalledState = func() (config.State, error) { return config.State{}, nil }
	suiteInstalled = func(name string) (bool, error) { return false, nil }
	loadVersionManifest = func() (config.VersionManifest, error) { return config.VersionManifest{}, nil }
	saveVersionManifest = func(manifest config.VersionManifest) error { return nil }
	addInstalledSuite = func(name string) error { return nil }
	checkConflicts = func(s Suite, installed []string) ([]PortConflict, error) { return []PortConflict{}, nil }
	pullCalls := 0
	pullImageWithRollback = func(ctx context.Context, image string, onProgress func(string)) error {
		pullCalls++
		if pullCalls == 2 {
			return errors.New("pull failed")
		}
		return nil
	}
	writeComposeFile = func(name string) (string, error) {
		t.Fatal("compose write should not run after pull failure")
		return "", nil
	}

	if err := Install(context.Background(), "boosting", InstallOptions{}); err == nil {
		t.Fatal("expected pull failure")
	}
}
