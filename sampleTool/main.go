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
	// This always shows
	log.Printf("Hello world, version %s, args %v", cli.ShortVersion, flag.Args())
	// This shows and is colorized and structured, unless loglevel is set to critical.
	log.S(log.Error, "Error test",
		log.Str("myflag", *myFlag),
		log.Attr("num_args", len(flag.Args())),
		log.Attr("args", flag.Args()))
}
