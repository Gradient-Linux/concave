package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type SetupStep string

const (
	StepHardwareDetection SetupStep = "hardware_detection"
	StepDriverInstall     SetupStep = "driver_install"
	StepSuiteSelection    SetupStep = "suite_selection"
	StepInternetCheck     SetupStep = "internet_check"
	StepImagePull         SetupStep = "image_pull"
	StepWorkspaceInit     SetupStep = "workspace_init"
	StepHealthCheck       SetupStep = "health_check"
)

type SetupState struct {
	CompletedSteps []SetupStep `json:"completed_steps"`
	LastRun        time.Time   `json:"last_run"`
	Complete       bool        `json:"complete"`
}

// LoadSetupState reads ~/gradient/config/setup.json or returns an empty state.
func LoadSetupState(workspaceRoot string) (SetupState, error) {
	path := filepath.Join(workspaceRoot, "config", "setup.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return SetupState{CompletedSteps: []SetupStep{}}, nil
		}
		return SetupState{}, fmt.Errorf("read %s: %w", path, err)
	}

	var state SetupState
	if err := json.Unmarshal(data, &state); err != nil {
		return SetupState{}, fmt.Errorf("unmarshal %s: %w", path, err)
	}
	if state.CompletedSteps == nil {
		state.CompletedSteps = []SetupStep{}
	}
	return state, nil
}

// SaveSetupState writes setup.json atomically.
func SaveSetupState(workspaceRoot string, state SetupState) error {
	if state.CompletedSteps == nil {
		state.CompletedSteps = []SetupStep{}
	}
	state.LastRun = time.Now().UTC()
	path := filepath.Join(workspaceRoot, "config", "setup.json")
	return writeSetupJSON(path, state)
}

// MarkStepComplete records a completed setup step.
func MarkStepComplete(workspaceRoot string, step SetupStep) error {
	state, err := LoadSetupState(workspaceRoot)
	if err != nil {
		return err
	}
	if !IsStepComplete(state, step) {
		state.CompletedSteps = append(state.CompletedSteps, step)
	}
	return SaveSetupState(workspaceRoot, state)
}

// MarkSetupComplete marks the setup flow as complete.
func MarkSetupComplete(workspaceRoot string) error {
	state, err := LoadSetupState(workspaceRoot)
	if err != nil {
		return err
	}
	state.Complete = true
	return SaveSetupState(workspaceRoot, state)
}

// IsStepComplete reports whether a setup step has already completed.
func IsStepComplete(state SetupState, step SetupStep) bool {
	for _, completed := range state.CompletedSteps {
		if completed == step {
			return true
		}
	}
	return false
}

func writeSetupJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename %s to %s: %w", tmp, path, err)
	}
	return nil
}
