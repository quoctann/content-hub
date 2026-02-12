package main

import (
	"github.com/quoctann/content-hub/cmd/app/commands"
)

// @title           Content Hub API
// @version         1.0
// @description     Content Hub API Server.

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        X-API-Key
func main() {
	// The API server logic is encapsulated in the commands and bootstrap packages.
	// We simply delegate execution to the root command.
	commands.Execute()
}
