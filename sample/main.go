package main

import (
	"os"

	"fortio.org/cli"
)

func main() { os.Exit(Main()) }

func Main() int {
	cli.Config.MinArgs = 2
	cli.Config.MaxArgs = 4
	if cli.ServerMain() {
		select {}
	}
	return 0
}
