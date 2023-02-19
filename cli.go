// Fortio CLI/Main utilities.
//
// (c) 2023 Fortio Authors
// See LICENSE

// Package cli contains utilities for command line tools and server main()s
// to handle flags, arguments, version, logging ([fortio.org/log]), etc...
// And for [ServerMain] the setup of a confimap/directory watch for flags
// and a config endpoint (uses [fortio.org/dflag]).
// Configure using the package variables (at minimum [MinArgs] unless your
// binary only accepts flags), setup additional [flag]s before calling
// [Main] or [ServerMain].
package cli // import "fortio.org/cli"

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"fortio.org/log"
	"fortio.org/version"
)

// Configuration for your Main() or ServerMain() function.
// These variables is how to setup the arguments, flags and usage parsing for [Main] and [ServerMain].
// At minium set the MinArgs should be set.
var (
	// *Version will be filled automatically by the cli package, using [fortio.org/version.FromBuildInfo()].
	ShortVersion string // x.y.z from tag/install
	LongVersion  string // version plus go version plus OS/arch
	FullVersion  string // LongVersion plus build date and git sha
	// Following can/should be specified.
	ProgramName string // Used at the beginning of Usage()
	// Cli usage/arguments example, ie "url1..." program name and "[flags]" will be added"
	// can include \n for additional details in the Usage() before the flags are dumped.
	ArgsHelp string
	MinArgs  int // Minimum number of arguments expected
	MaxArgs  int // Maximum number of arguments expected. 0 means same as MinArgs. -1 means no limit.
	// If not set to true, will setup static loglevel flag and logger output for client tools.
	ServerMode = false
	// Override this to change the exit function (for testing), will be applied to log.Fatalf too.
	ExitFunction = os.Exit
	baseExe      string
)

func Usage(w io.Writer, msg string, args ...any) {
	_, _ = fmt.Fprintf(w, "%s %s usage:\n\t%s [flags]%s\nor 1 of the special arguments\n\t%s {help|version|buildinfo}\nflags:\n",
		ProgramName,
		ShortVersion,
		baseExe,
		ArgsHelp,
		os.Args[0],
	)
	flag.CommandLine.SetOutput(w)
	flag.PrintDefaults()
	if msg != "" {
		fmt.Fprintf(w, msg, args...)
		fmt.Fprintln(w)
	}
}

// Show usage and error message on stderr and exit with code 1 or returns false.
func ErrUsage(msg string, args ...any) bool {
	Usage(os.Stderr, msg, args...)
	ExitFunction(1)
	// not reached, typically
	return false
}

// Main handles your commandline and flag parsing. Sets up flags first then call Main.
// For a server with dynamic flags, call ServerMain instead.
// Will either have called ExitFunction() (defaults to [os.Exit])
// or returned if all validations passed.
func Main() {
	quietFlag := flag.Bool("quiet", false, "Quiet mode, sets log level to warning")
	ShortVersion, LongVersion, FullVersion = version.FromBuildInfo()
	log.Config.FatalExit = ExitFunction
	baseExe = filepath.Base(os.Args[0])
	if ProgramName == "" {
		ProgramName = baseExe
	}
	if MaxArgs == 0 {
		MaxArgs = MinArgs
	}
	if ArgsHelp == "" {
		for i := 1; i <= MinArgs; i++ {
			ArgsHelp += fmt.Sprintf(" arg%d", i)
		}
		if MaxArgs < 0 {
			ArgsHelp += " ..."
		} else if MaxArgs > MinArgs {
			ArgsHelp += fmt.Sprintf(" [arg%d...arg%d]", MinArgs+1, MaxArgs)
		}
	}
	// Callers can pass that part of the help with or without leading space
	if !strings.HasPrefix(ArgsHelp, " ") {
		ArgsHelp = " " + ArgsHelp
	}
	if !ServerMode {
		log.SetDefaultsForClientTools()
		log.LoggerStaticFlagSetup("loglevel")
	}
	flag.CommandLine.Usage = func() { Usage(os.Stderr, "") } // flag handling will exit 1 after calling usage, except for -h/-help
	flag.Parse()
	nArgs := len(flag.Args())
	if nArgs == 1 {
		switch strings.ToLower(flag.Arg(0)) {
		case "version":
			fmt.Println(ShortVersion)
			ExitFunction(0)
			return // not typically reached, unless ExitFunction doesn't exit
		case "buildinfo":
			fmt.Print(FullVersion)
			ExitFunction(0)
			return // not typically reached, unless ExitFunction doesn't exit
		case "help":
			Usage(os.Stdout, "")
			ExitFunction(0)
			return // not typically reached, unless ExitFunction doesn't exit
		}
	}
	argsRange := (MinArgs != MaxArgs)
	exactly := "Exactly"
	if nArgs < MinArgs {
		if argsRange {
			exactly = "At least"
		}
		ErrArgCount(exactly, MinArgs, nArgs)
		return // not typically reached, unless ExitFunction doesn't exit
	}
	if MaxArgs >= 0 && nArgs > MaxArgs {
		if MaxArgs <= 0 {
			ErrUsage("No arguments expected (except for version, buildinfo or help and -flags), got %d", nArgs)
			return // not typically reached, unless ExitFunction doesn't exit
		}
		if argsRange {
			exactly = "At most"
		}
		ErrArgCount(exactly, MaxArgs, nArgs)
		return // not typically reached, unless ExitFunction doesn't exit
	}
	if *quietFlag {
		log.SetLogLevelQuiet(log.Warning)
	}
}

func ErrArgCount(prefix string, expected, actual int) bool {
	return ErrUsage("%s %d %s expected, got %d", prefix, expected, Plural(expected, "argument"), actual)
}

// Plural adds an "s" to the noun if i is not 1.
func Plural(i int, noun string) string {
	return PluralExt(i, noun, "s")
}

// PluralExt returns the noun with an extension if i is not 1.
// Eg:
//
//	PluralExt(1, "address", "es") // -> "address"
//	PluralExt(3 /* or 0 */, "address", "es") // -> "addresses"
func PluralExt(i int, noun string, ext string) string {
	if i == 1 {
		return noun
	}
	return noun + ext
}
