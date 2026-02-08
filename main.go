package main

import (
	"embed"

	"github.com/khang/google-suite-cli/cmd"
)

//go:embed all:skills/gsuite-manager
var skillFS embed.FS

func main() {
	cmd.SkillFS = skillFS
	cmd.Execute()
}
