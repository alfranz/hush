package main

import (
	"runtime/debug"

	"github.com/alfranz/hush/internal/cli"
)

// version is set at release time: -ldflags "-X main.version=1.2.3".
var version = ""

func main() {
	cli.Execute(resolveVersion())
}

// resolveVersion returns the ldflags version, or the module version for
// builds from "go install github.com/alfranz/hush/cmd/hush@v1.2.3".
func resolveVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
