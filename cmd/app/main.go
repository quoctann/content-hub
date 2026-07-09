package main

import (
	"github.com/quoctann/content-hub/cmd/app/commands"
)

// @title           Content Hub API
// @version         1.0
// @description     Content Hub API Server.

// @host      localhost:8080
// @BasePath  /
func main() {
	commands.Execute()
}
