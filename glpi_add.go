package main

import (
	"os"

	glpi_cmd "github.com/kjuchniewicz/glpi_add/cmd"
)

func main() {
	cmd := glpi_cmd.NewRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
