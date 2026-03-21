package main

import "github.com/Gradient-Linux/concave/cmd"

// Version is injected at build time.
var Version = "dev"

var executeCommand = cmd.Execute

func run(version string) {
	cmd.Version = version
	executeCommand()
}

func main() {
	run(Version)
}
