package main

import (
	"github.com/hchenc/reloader/cmd/app"
	"os"
)

func main() {
	cmd := app.NewReloaderCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
