package main

import (
	"github.com/quoctann/content-hub/cmd/app/commands"
)

func main() {
	// The API server logic is encapsulated in the commands and bootstrap packages.
	// We simply delegate execution to the root command.
	commands.Execute()
}
