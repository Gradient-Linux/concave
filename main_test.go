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
