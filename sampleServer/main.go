package main

import (
	"fortio.org/cli"
)

func main() {
	cli.MinArgs = 2
	cli.MaxArgs = 4
	if !cli.ServerMain() {
		// in reality in both case we'd start some actual server
		return
	}
	select {}
}
