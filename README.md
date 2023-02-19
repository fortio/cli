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

func main() {
	myFlag := flag.String("myflag", "default", "my flag")
	cli.MinArgs = 2
	cli.MaxArgs = 4
	cli.Main() // Will have either called cli.ExitFunction or everything is valid
	// Next line output won't show when passed -quiet
	log.Infof("Info test, -myflag is %q", *myFlag)
	log.Printf("Hello world, version %s, args %v", cli.ShortVersion, flag.Args())
}
```

```bash
$ sampleTool a
sampleTool 1.0.0 usage:
	sampleTool [flags] arg1 arg2 [arg3...arg4]
or 1 of the special arguments
	sampleTool {help|version|buildinfo}
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

## Additional builtins

### buildinfo

e.g

```bash
$ go install fortio.org/cli/sampleServer@latest
go: downloading fortio.org/cli v0.1.0
$ sampleServer buildinfo
0.1.0 h1:M+coC8So/41xSyGiiJ/6RS+XhnshNUslZuaB6H8z9GI= go1.19.6 arm64 darwin
go	go1.19.6
path	fortio.org/cli/sampleServer
mod	fortio.org/cli	v0.1.0	h1:M+coC8So/41xSyGiiJ/6RS+XhnshNUslZuaB6H8z9GI=
dep	fortio.org/dflag	v1.4.1	h1:WDhlHMh3yrQFrvspyN5YEyr8WATdKM2dUJlTxsjCDtI=
dep	fortio.org/fortio	v1.50.1	h1:5FSttAHQsyAsi3dzxDmSByfzDYByrWY/yw53bqOg+Kc=
dep	fortio.org/log	v1.2.2	h1:vs42JjNwiqbMbacittZjJE9+oi72Za6aekML9gKmILg=
dep	fortio.org/version	v1.0.2	h1:8NwxdX58aoeKx7T5xAPO0xlUu1Hpk42nRz5s6e6eKZ0=
dep	github.com/fsnotify/fsnotify	v1.6.0	h1:n+5WquG0fcWoWp6xPWfHdbskMCQaFnG6PfBrh1Ky4HY=
dep	github.com/google/uuid	v1.3.0	h1:t6JiXgmwXMjEs8VusXIJk2BXHsn+wx8BZdTaoZ5fu7I=
dep	golang.org/x/exp	v0.0.0-20230213192124-5e25df0256eb	h1:PaBZQdo+iSDyHT053FjUCgZQ/9uqVwPOcl7KSWhKn6w=
dep	golang.org/x/net	v0.7.0	h1:rJrUqqhjsgNp7KqAIc25s9pZnjU7TUcSY7HcVZjdn1g=
dep	golang.org/x/sys	v0.5.0	h1:MUK/U/4lj1t1oPg0HfuXDN/Z1wv31ZJ/YcPiGccS4DU=
dep	golang.org/x/text	v0.7.0	h1:4BRB4x83lYWy72KwLD/qYDuTu7q9PjSagHvijDw7cLo=
build	-compiler=gc
build	CGO_ENABLED=1
build	CGO_CFLAGS=
build	CGO_CPPFLAGS=
build	CGO_CXXFLAGS=
build	CGO_LDFLAGS=
build	GOARCH=arm64
build	GOOS=darwin
```

### help
```bash
$ sampleServer help
sampleServer 0.1.0 usage:
	sampleServer [flags] arg1 arg2 [arg3...arg4]
or 1 of the special arguments
	sampleServer {help|version|buildinfo}
flags:
  -config-dir directory
    	Config directory to watch for dynamic flag changes
  -config-port port
    	Config port to open for dynamic flag UI/api
  -loglevel level
    	log level, one of [Debug Verbose Info Warning Error Critical Fatal] (default Info)
  -quiet
    	Quiet mode, sets log level to warning
```

### version
Short 'numeric' version (v skipped, useful for docker image tags etc)
```bash
$ sampleServer version
0.1.1
```
