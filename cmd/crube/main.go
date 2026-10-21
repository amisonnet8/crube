// Command crube shows and keeps descriptions of the files and directories of a
// project (docs/reference/cli.md). Everything is in internal/cli.
package main

import (
	"os"

	"github.com/amisonnet8/crube/internal/cli"
)

// version is set by the release build (-ldflags "-X main.version=…").
var version string

func main() {
	os.Exit(cli.Run(os.Args[1:], cli.Env{
		Getenv:  os.Getenv,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Version: version,
	}))
}
