package main

import (
	"github.com/G-core/gcore-cli/internal/commands"
	"github.com/G-core/gcore-cli/internal/core"
)

func main() {
	core.Execute(commands.Commands())
}
