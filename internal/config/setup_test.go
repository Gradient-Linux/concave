package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetupStateRoundTripAndIdempotency(t *testing.T) {
	root := t.TempDir()

	state, err := LoadSetupState(root)
	if err != nil {
		t.Fatalf("LoadSetupState() error = %v", err)
	}
	if len(state.CompletedSteps) != 0 {
		t.Fatalf("unexpected initial state: %#v", state)
	}

	if err := MarkStepComplete(root, StepHardwareDetection); err != nil {
		t.Fatalf("MarkStepComplete() error = %v", err)
	}
	if err := MarkStepComplete(root, StepHardwareDetection); err != nil {
		t.Fatalf("MarkStepComplete() idempotent error = %v", err)
	}
	if err := MarkSetupComplete(root); err != nil {
		t.Fatalf("MarkSetupComplete() error = %v", err)
	}

	state, err = LoadSetupState(root)
	if err != nil {
		t.Fatalf("LoadSetupState() second error = %v", err)
	}
	if !IsStepComplete(state, StepHardwareDetection) {
		t.Fatalf("expected %q to be complete: %#v", StepHardwareDetection, state)
	}
	if !state.Complete {
		t.Fatalf("expected complete state: %#v", state)
	}
	if len(state.CompletedSteps) != 1 {
		t.Fatalf("step count = %d, want 1", len(state.CompletedSteps))
	}
}

func TestSaveSetupStateWritesAtomically(t *testing.T) {
	root := t.TempDir()
	state := SetupState{CompletedSteps: []SetupStep{StepInternetCheck}}
	if err := SaveSetupState(root, state); err != nil {
		t.Fatalf("SaveSetupState() error = %v", err)
	}

	path := filepath.Join(root, "config", "setup.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected setup.json: %v", err)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("unexpected temp file state: %v", err)
	}
}
