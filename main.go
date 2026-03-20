package main

import "github.com/gradient-linux/concave/cmd"

// Version is injected at build time.
var Version = "dev"

func main() {
	cmd.Version = Version
	cmd.Execute()
}
