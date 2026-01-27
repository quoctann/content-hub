package main

import (
	"github.com/quoctann/content-hub/cmd/db/commands"
)

func main() {
	// The database CLI logic is encapsulated in the commands and bootstrap packages.
	// We simply delegate execution to the commands package.
	commands.Execute()
}
