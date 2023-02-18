# cli
Reduce boiler plate needed on each new Golang main functions (Command Line Interface) for both tool and servers

It abstracts the repetitive parts of a `main()` command line tool, flag parsing, usage, etc...


## Tool Example
Client/Tool example (no dynamic flag url or config) [sampleTool](sampleTool/main.go)

Code as simple as
```golang
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
```

```bash
$ sampleTool a
sampleTool 1.0.0 usage:
	sampleTool [flags] arg1 arg2 [arg3...arg4]
flags:
  -loglevel level
    	log level, one of [Debug Verbose Info Warning Error Critical Fatal] (default Info)
  -myflag string
    	my flag (default "default")
  -quiet
    	Quiet mode, sets log level to warning
At least 2 arguments expected, got 1
```

or normal case
```bash
$ sampleTool a b
15:42:17 I Info test, -myflag is "default"
15:42:17 Hello world, version dev, args [a b]
```

## Server Example

Server example [sampleServer](sampleServer/main.go)


```bash
% go run . -config-dir ./config -config-port 8888 a b
15:45:20 I updater.go:47> Configmap flag value watching on ./config
15:45:20 I updater.go:156> updating loglevel to "verbose\n"
15:45:20 I logger.go:183> Log level is now 1 Verbose (was 2 Info)
15:45:20 I updater.go:97> Now watching . and config
15:45:20 I updater.go:162> Background thread watching config now running
15:45:20 Fortio 1.50.1 config server listening on tcp [::]:8888
15:45:20 Starting sampleServer dev  go1.19.6 arm64 darwin
# When visiting the UI
15:46:20 ListFlags: GET / HTTP/1.1 [::1]:52406 ()  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
```

With the flags ui on http://localhost:8888

<img width="716" alt="flags UI" src="https://user-images.githubusercontent.com/3664595/219904547-368a024e-1d6a-4301-a7a9-8882e37f5a90.png">
