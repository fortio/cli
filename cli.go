// Fortio CLI/Main utilities.
//
// (c) 2023 Fortio Authors
// See LICENSE

// Package cli contains utilities for command line tools and server main()s
// to handle flags, arguments, version, logging ([fortio.org/log]), etc...
// And for [ServerMain] the setup of a confimap/directory watch for flags
// and a config endpoint (uses [fortio.org/dflag]).
// Configure [Config] ([Configuration] fields), setup additional [flag]s
// before calling [Main] or [ServerMain].
package cli // import "fortio.org/cli"

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"fortio.org/dflag/configmap"
	"fortio.org/dflag/dynloglevel"
	"fortio.org/dflag/endpoint"
	"fortio.org/fortio/fhttp"
	"fortio.org/log"
	"fortio.org/version"
)

// Configuration for your Main() or ServerMain() function.
type Configuration struct {
	// Will be filled automatically by the cli package, using fortio.org/version FromBuildInfo().
	ShortVersion string // x.y.z from tag/install
	LongVersion  string // version plus go version plus OS/arch
	FullVersion  string // LongVersion plus build date and git sha
	// Following should be specified.
	ProgramName string // Used at the beginning of Usage()
	// Cli usage/arguments example, ie "url1..." program name and "[flags]" will be added"
	// can include \n for additional details in the Usage() before the flags are dumped.
	ArgsHelp string
	MinArgs  int // Minimum number of arguments expected
	MaxArgs  int // Maximum number of arguments expected. 0 means same as MinArgs. -1 means no limit.
}

var (
	// Config is how to setup the arguments, flags and usage parsing for [Main] and [ServerMain].
	// At minium set the MinArgs and MaxArgs fields. See [Configuration] for the fields.
	Config    Configuration
	QuietFlag = flag.Bool("quiet", false, "Quiet mode, sets log level to warning")
	// If not set to true, will setup static loglevel flag and logger output for client tools.
	ServerMode = false
	// Override this to change the exit function (for testing), will be applied to log.Fatalf too.
	ExitFunction = os.Exit
)

func Usage(w io.Writer, msg string, args ...any) {
	_, _ = fmt.Fprintf(w, "%s %s usage:\n\t%s [flags]%s\nflags:\n",
		Config.ProgramName,
		Config.ShortVersion,
		os.Args[0],
		Config.ArgsHelp)
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
// returns true if there was no error.
// Might not return and have already exited for help/usage/etc...
func Main() bool {
	Config.ShortVersion, Config.LongVersion, Config.FullVersion = version.FromBuildInfo()
	log.Config.FatalExit = ExitFunction
	if Config.ProgramName == "" {
		Config.ProgramName = filepath.Base(os.Args[0])
	}
	if Config.MaxArgs == 0 {
		Config.MaxArgs = Config.MinArgs
	}
	if Config.ArgsHelp == "" {
		for i := 1; i <= Config.MinArgs; i++ {
			Config.ArgsHelp += fmt.Sprintf(" arg%d", i)
		}
		if Config.MaxArgs < 0 {
			Config.ArgsHelp += " ..."
		} else if Config.MaxArgs > Config.MinArgs {
			Config.ArgsHelp += fmt.Sprintf(" [arg%d...arg%d]", Config.MinArgs+1, Config.MaxArgs)
		}
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
			fmt.Println(Config.ShortVersion)
			ExitFunction(0)
		case "buildinfo":
			fmt.Print(Config.FullVersion)
			ExitFunction(0)
		case "help":
			Usage(os.Stdout, "")
			ExitFunction(0)
		}
	}
	if nArgs < Config.MinArgs {
		return ErrUsage("At least %d arguments expected, got %d", Config.MinArgs, nArgs)
	}
	if Config.MaxArgs >= 0 && nArgs > Config.MaxArgs {
		if Config.MaxArgs <= 0 {
			return ErrUsage("No arguments expected (except for version, buildinfo or help and -flags), got %d", nArgs)
		}
		return ErrUsage("At most %d arguments expected, got %d", Config.MaxArgs, nArgs)
	}
	if *QuietFlag {
		log.SetLogLevelQuiet(log.Warning)
	}
	return true
}

// ServerMain returns true if a config port server has been started
// caller needs to select {} after its own code is ready.
// Will have exited if there are usage errors (wrong number of arguments, bad flags etc...).
// It sets up (optional) config-dir to watch and listen on config-port for dynamic flag
// changes and UI/api.
func ServerMain() bool {
	ConfigDir := flag.String("config-dir", "", "Config `directory` to watch for dynamic flag changes")
	ConfigPort := flag.String("config-port", "", "Config `port` to open for dynamic flag UI/api")
	dynloglevel.LoggerFlagSetup("loglevel")
	ServerMode = true
	if !Main() {
		return false
	}

	if *ConfigDir != "" {
		if _, err := configmap.Setup(flag.CommandLine, *ConfigDir); err != nil {
			log.Critf("Unable to watch config/flag changes in %v: %v", *ConfigDir, err)
		}
	}
	hasStartedServer := false
	if *ConfigPort != "" {
		mux, addr := fhttp.HTTPServer("config", *ConfigPort) // err already logged
		if addr != nil {
			hasStartedServer = true
			setURL := "/set"
			ep := endpoint.NewFlagsEndpoint(flag.CommandLine, setURL)
			mux.HandleFunc("/", ep.ListFlags)
			mux.HandleFunc(setURL, ep.SetFlag)
		}
	}
	log.Printf("Starting %s %s", Config.ProgramName, Config.LongVersion)
	return hasStartedServer
}
