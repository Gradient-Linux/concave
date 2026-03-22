package main

import (
	"os"

	"github.com/Gradient-Linux/concave/cmd"
	"github.com/Gradient-Linux/concave/internal/system"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

var executeCommand = cmd.Execute

func run(version string) {
	cmd.Version = version
	cmd.Commit = Commit
	cmd.BuildDate = BuildDate
	executeCommand()
}

func main() {
	defer system.InstallCrashHandler(Version, os.Args)
	run(Version)
}
