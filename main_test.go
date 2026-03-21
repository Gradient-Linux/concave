package main

import (
	"testing"

	"github.com/Gradient-Linux/concave/cmd"
)

func TestRunSetsCommandVersion(t *testing.T) {
	previousExecute := executeCommand
	previousVersion := cmd.Version
	t.Cleanup(func() {
		executeCommand = previousExecute
		cmd.Version = previousVersion
	})

	called := false
	executeCommand = func() {
		called = true
	}

	run("v1.2.3")

	if !called {
		t.Fatal("expected executeCommand to be invoked")
	}
	if cmd.Version != "v1.2.3" {
		t.Fatalf("cmd.Version = %q", cmd.Version)
	}
}

func TestMainUsesInjectedVersion(t *testing.T) {
	previousExecute := executeCommand
	previousCmdVersion := cmd.Version
	previousMainVersion := Version
	t.Cleanup(func() {
		executeCommand = previousExecute
		cmd.Version = previousCmdVersion
		Version = previousMainVersion
	})

	called := false
	executeCommand = func() {
		called = true
	}
	Version = "v9.9.9"

	main()

	if !called {
		t.Fatal("expected main to invoke executeCommand")
	}
	if cmd.Version != "v9.9.9" {
		t.Fatalf("cmd.Version = %q", cmd.Version)
	}
}
