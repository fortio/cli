package main

import (
	"flag"

	"fortio.org/cli"
	"fortio.org/log"
)

func main() {
	myFlag := flag.String("myflag", "default", "my flag")
	cli.MinArgs = 2
	cli.MaxArgs = 4
	cli.Main() // Will have either called cli.ExitFunction or everything is valid
	// Next line output won't show when passed -quiet
	log.Infof("Info test, -myflag is %q", *myFlag)
	log.Printf("Hello world, version %s, args %v", cli.ShortVersion, flag.Args())
}
