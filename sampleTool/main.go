package main

import (
	"flag"
	"os"

	"fortio.org/cli"
	"fortio.org/log"
)

func main() { os.Exit(Main()) }

func Main() int {
	myFlag := flag.String("myflag", "default", "my flag")
	cli.Config.MinArgs = 2
	cli.Config.MaxArgs = 4
	if !cli.Main() {
		return 1
	}
	log.Infof("Info test, -myflag is %q", *myFlag) // won't show with -quiet
	log.Printf("Hello world, version %s, args %v", cli.Config.ShortVersion, flag.Args())
	return 0
}
