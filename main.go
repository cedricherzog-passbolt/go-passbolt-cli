package main

import (
	"github.com/passbolt/go-passbolt-cli/internal/cmd"
)

func main() {
	cmd.SetVersionInfo(version, commit, date, dirty)
	cmd.Execute()
}
