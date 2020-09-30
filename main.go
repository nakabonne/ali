package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/k0kubun/pp"
	flag "github.com/spf13/pflag"

	"github.com/nakabonne/ali/gui"
)

var (
	flagSet = flag.NewFlagSet("ali", flag.ContinueOnError)

	usage = func() {
		fmt.Fprintln(os.Stderr, `Usage:
  ali [flags] <target URL>

Flags:`)
		flagSet.PrintDefaults()
		fmt.Println("")
		fmt.Fprintln(os.Stderr, `Examples:
  ali --duration=10m --rate=100 http://exmple.com`)
		fmt.Println("")
		fmt.Fprintln(os.Stderr, `Author:
  Ryo Nakao <ryo@nakao.dev>`)
	}
	// Automatically populated by goreleaser during build
	version = "unversioned"
	commit  = "?"
	date    = "?"
)

type cli struct {
	target   string
	rate     int
	duration time.Duration
	timeout  time.Duration
	method   string
	header   string
	body     string

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
	flagSet.IntVarP(&c.rate, "rate", "r", 50, "the request rate per second to issue against the targets")
	flagSet.DurationVarP(&c.duration, "duration", "d", time.Second*10, "the amount of time to issue requests to the targets")
	flagSet.DurationVarP(&c.timeout, "timeout", "t", time.Second*30, "the timeout for each request")
	flagSet.StringVarP(&c.method, "method", "m", "GET", "an HTTP request method for each request")
	flagSet.StringVarP(&c.header, "header", "H", "", "a request header to be sent")
	flagSet.StringVarP(&c.body, "body", "b", "", "the file whose content will be set as the http request body")
	flagSet.BoolVarP(&c.version, "version", "v", false, "print the current version")
	flagSet.BoolVar(&c.debug, "debug", false, "run in debug mode")
	flagSet.Usage = usage
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(c.stderr, err)
		}
		return
	}

	os.Exit(c.run(flagSet.Args()))
}

func (c *cli) run(args []string) int {
	if c.version {
		fmt.Fprintf(c.stderr, "version=%s, commit=%s, buildDate=%s, os=%s, arch=%s\n", version, commit, date, runtime.GOOS, runtime.GOARCH)
		return 0
	}
	if len(args) == 0 {
		fmt.Fprintln(c.stderr, "no target given")
		usage()
		return 1
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
		return
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
