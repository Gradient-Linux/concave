package main

import "github.com/gradientlinux/concave/cmd"

// Version is injected at build time.
var Version = "dev"

func main() {
	cmd.Version = Version
	cmd.Execute()
}
