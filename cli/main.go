package main

import (
	"os"

	"github.com/avarian/primbon-ajaib-backend/cli/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
