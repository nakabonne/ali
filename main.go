package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/k0kubun/pp"
	flag "github.com/spf13/pflag"

	"github.com/nakabonne/ali/gui"
)

var (
	flagSet = flag.NewFlagSet("ali", flag.ContinueOnError)

	usage = func() {
		fmt.Fprintln(os.Stderr, "usage: ali [<flag> ...]")
		flagSet.PrintDefaults()
	}
	// Automatically populated by goreleaser during build
	version = "unversioned"
	commit  = "?"
	date    = "?"
)

type cli struct {
	debug   bool
	version bool
	stdout  io.Writer
	stderr  io.Writer
}

func main() {
	c := &cli{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	flagSet.BoolVarP(&c.version, "version", "v", false, "print the current version")
	flagSet.BoolVar(&c.debug, "debug", false, "run in debug mode")
	flagSet.Usage = usage
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(c.stderr, err)
		}
		return
	}

	os.Exit(c.run())
}

func (c *cli) run() int {
	if c.version {
		fmt.Fprintf(c.stderr, "version=%s, commit=%s, buildDate=%s, os=%s, arch=%s\n", version, commit, date, runtime.GOOS, runtime.GOARCH)
		return 0
	}
	setDebug(nil, c.debug)

	if err := gui.Run(); err != nil {
		fmt.Fprintf(c.stderr, "failed to start application: %s", err.Error())
		return 1
	}

	return 0
}

// Makes a new file under the working directory only when debug use.
func setDebug(w io.Writer, debug bool) {
	if !debug {
		pp.SetDefaultOutput(ioutil.Discard)
	}
	if w == nil {
		var err error
		w, err = os.OpenFile(filepath.Join(".", "ali-debug.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
	}
	pp.SetDefaultOutput(w)
}
