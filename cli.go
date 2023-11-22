// Fortio CLI/Main utilities.
//
// (c) 2023 Fortio Authors
// See LICENSE

// Package cli contains utilities for command line tools and server main()s
// to handle flags, arguments, version, logging ([fortio.org/log]), etc...
// Configure using the package variables (at minimum [MinArgs] unless your
// binary only accepts flags), setup additional [flag] before calling
// [Main] or [fortio.org/scli.ServerMain] for configmap and dynamic flags
// setup.
// Also supports (sub)commands style, see [CommandBeforeFlags] and [Command].
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
	// Out parameters:
	// *Version will be filled automatically by the cli package, using [fortio.org/version.FromBuildInfo()].
	ShortVersion string // x.y.z from tag/install
	LongVersion  string // version plus go version plus OS/arch
	FullVersion  string // LongVersion plus build date and git sha
	Command      string // first argument, if [CommandBeforeFlags] is true.
	// Following can/should be specified.
	ProgramName string // Used at the beginning of Usage()
	// Optional for programs using subcommand, command will be set in [Command].
	// If you wish to replace the help default colorize `command` with something else set CommandHelp.
	CommandBeforeFlags bool
	// Cli usage/arguments example, ie "url1..." program name and "[flags]" will be added"
	// can include \n for additional details in the Usage() before the flags are dumped.
	ArgsHelp string
	// Command help will be used instead of purple "command " in help text for cli that have a
	// command before the flags (when [CommandBeforeFlags] is true). For instance you could use
	// cli.CommandHelp = "{" + cli.ColorJoin(log.Colors.Purple, "a", "b", "c") + "}"
	// for colorize {a|b|c} in the help before [flags].
	CommandHelp string
	MinArgs     int // Minimum number of arguments expected, not counting (optional) command.
	MaxArgs     int // Maximum number of arguments expected. 0 means same as MinArgs. -1 means no limit.
	// If not set to true, will setup static loglevel flag and logger output for client tools.
	ServerMode = false
	// Override this to change the exit function (for testing), will be applied to log.Fatalf too.
	ExitFunction = os.Exit
	// Hook to call before flag.Parse() - for instance to use ChangeFlagDefaults for logger flags etc.
	BeforeFlagParseHook = func() {}
	// Calculated base exe name from args (will be used if ProgramName if not set).
	baseExe string
	// List of functions to call for env help
	EnvHelpFuncs []func(w io.Writer) = []func(w io.Writer){log.EnvHelp}
)

// ChangeFlagsDefault sets some flags to a different default.
// Will return error if the flag is not found and value can't be set
// (caller should likely log.Fatalf if that happens as it's a typo/bug).
func ChangeFlagsDefault(newDefault string, flagNames ...string) error {
	for _, flagName := range flagNames {
		f := flag.Lookup(flagName)
		if f == nil {
			return fmt.Errorf("flag %q not found", flagName)
		}
		f.DefValue = newDefault
		err := f.Value.Set(newDefault)
		if err != nil {
			return err
		}
	}
	return nil
}

func ColorJoin(color string, args ...string) string {
	return color + strings.Join(args, log.Colors.Reset+"|"+color) + log.Colors.Reset
}

func usage(w io.Writer, msg string, args ...any) {
	cmd := ""
	if CommandBeforeFlags {
		if CommandHelp != "" {
			cmd = CommandHelp + " "
		} else {
			cmd = log.Colors.Purple + "command " + log.Colors.Reset
		}
	}
	_, _ = fmt.Fprintf(w, log.Colors.Reset+"%s %s usage:\n\t%s %s["+
		log.Colors.Cyan+"flags"+log.Colors.Reset+"]%s\nor 1 of the special arguments\n\t%s {"+
		ColorJoin(log.Colors.Purple, "help", "envhelp", "version", "buildinfo")+"}\n"+"flags:\n"+log.Colors.Cyan,
		ProgramName,
		log.Colors.Blue+ShortVersion+log.Colors.Reset,
		baseExe,
		cmd,
		ArgsHelp,
		os.Args[0],
	)
	var buf strings.Builder
	flag.CommandLine.SetOutput(&buf)
	flag.PrintDefaults()
	flags := buf.String()
	flags = strings.ReplaceAll(flags, "(default ", log.Colors.Green+"(default ")
	flags = strings.ReplaceAll(flags, "\n", "\n"+log.Colors.Cyan)
	flags = strings.ReplaceAll(flags, "\t", log.Colors.Reset+"\t")
	_, _ = w.Write([]byte(flags + log.Colors.Reset))
	if msg != "" {
		fmt.Fprintf(w, msg, args...)
		fmt.Fprintln(w)
	}
}

// Main handles your commandline and flag parsing. Sets up flags first then call Main.
// For a server with dynamic flags, call ServerMain instead.
// Will either have called [ExitFunction] (defaults to [os.Exit])
// or returned if all validations passed.
func Main() {
	quietFlag := flag.Bool("quiet", false,
		"Quiet mode, sets loglevel to Error (quietly) to reduces the output")
	flag.BoolVar(&log.Config.ForceColor, "logger-force-color", false,
		"Force color output even if stderr isn't a terminal")
	nocolor := flag.Bool("logger-no-color", false,
		"Prevent colorized output even if stderr is a terminal")
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
	flag.CommandLine.Usage = func() { usage(os.Stderr, "") } // flag handling will exit 1 after calling usage, except for -h/-help
	BeforeFlagParseHook()
	nArgs := len(os.Args)
	if nArgs == 2 {
		switch strings.ToLower(os.Args[1]) {
		case "version":
			fmt.Println(ShortVersion)
			ExitFunction(0)
			return // not typically reached, unless ExitFunction doesn't exit
		case "buildinfo":
			fmt.Print(FullVersion)
			ExitFunction(0)
			return // not typically reached, unless ExitFunction doesn't exit
		case "help":
			usage(os.Stdout, "")
			ExitFunction(0)
			return // not typically reached, unless ExitFunction doesn't exit
		case "envhelp":
			fmt.Println("# Environment variables recognized and current values:")
			for _, f := range EnvHelpFuncs {
				f(os.Stdout)
			}
			ExitFunction(0)
			return // not typically reached, unless ExitFunction doesn't exit
		}
	}
	if CommandBeforeFlags {
		if nArgs == 1 {
			ErrUsage("Missing command argument")
			return // not typically reached, unless ExitFunction doesn't exit
		}
		Command = os.Args[1]
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
	}
	// In case of a bad flag, we want it in red when on console:
	os.Stderr.WriteString(log.Colors.BrightRed)
	flag.Parse()
	os.Stderr.WriteString(log.Colors.Reset)
	if *nocolor {
		// Don't override the env if the flag isn't set
		// (downside is if LOGGER_FORCE_COLOR is set to false, this -logger-no-color=false can't override it)
		log.Config.ForceColor = !*nocolor
	}
	log.SetColorMode()
	nArgs = len(flag.Args())
	argsRange := (MinArgs != MaxArgs)
	exactly := "Exactly"
	if nArgs < MinArgs {
		if argsRange {
			exactly = "At least"
		}
		errArgCount(exactly, MinArgs, nArgs)
		return // not typically reached, unless ExitFunction doesn't exit
	}
	if MaxArgs >= 0 && nArgs > MaxArgs {
		if MaxArgs <= 0 {
			ErrUsage("No arguments expected (except for version, buildinfo, help or envhelp and -flags), got %d", nArgs)
			return // not typically reached, unless ExitFunction doesn't exit
		}
		if argsRange {
			exactly = "At most"
		}
		errArgCount(exactly, MaxArgs, nArgs)
		return // not typically reached, unless ExitFunction doesn't exit
	}
	if *quietFlag {
		log.SetLogLevelQuiet(log.Error)
	}
}

func errArgCount(prefix string, expected, actual int) {
	ErrUsage("%s %d %s expected, got %d", prefix, expected, Plural(expected, "argument"), actual)
}

// Show usage and error message on stderr and calls [ExitFunction] with code 1.
func ErrUsage(msg string, args ...any) {
	usage(os.Stderr, log.Colors.BrightRed+msg+log.Colors.Reset, args...)
	ExitFunction(1)
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
