package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/k0kubun/pp"
	flag "github.com/spf13/pflag"

	"github.com/nakabonne/ali/attacker"
	"github.com/nakabonne/ali/gui"
)

var (
	flagSet = flag.NewFlagSet("ali", flag.ContinueOnError)

	usage = func() {
		fmt.Fprintln(os.Stderr, `Usage:
  ali [flags] <target URL>

Flags:`)
		flagSet.PrintDefaults()
		fmt.Fprintln(os.Stderr, `
Examples:
  ali --duration=10m --rate=100 http://host.xz

Author:
  Ryo Nakao <ryo@nakao.dev>
`)
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
	bodyFile string

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
	flagSet.StringVarP(&c.body, "body", "b", "", "a request body to be sent")
	flagSet.StringVarP(&c.bodyFile, "body-file", "B", "", "the file whose content will be set as the http request body")
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
	target := args[0]
	if _, err := url.ParseRequestURI(target); err != nil {
		fmt.Fprintf(c.stderr, "bad target URL: %v\n", err)
		usage()
		return 1
	}
	opts, err := c.makeOptions()
	if err != nil {
		fmt.Fprintln(c.stderr, err.Error())
		usage()
		return 1
	}
	setDebug(nil, c.debug)
	if err := gui.Run(target, opts); err != nil {
		fmt.Fprintf(c.stderr, "failed to start application: %s\n", err.Error())
		usage()
		return 1
	}

	return 0
}

// makeOptions gives back an options for attacker, with the CLI input.
func (c *cli) makeOptions() (*attacker.Options, error) {
	if !validateMethod(c.method) {
		return nil, fmt.Errorf("given method %q isn't an HTTP request method", c.method)
	}

	header := make(http.Header)
	if c.header != "" {
		parts := strings.SplitN(c.header, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("given header %q has a wrong format", c.header)
		}
		key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if key == "" || val == "" {
			return nil, fmt.Errorf("given header %q has a wrong format", c.header)
		}
		// NOTE: Add key/value directly to the http.Header (map[string][]string).
		// http.Header.Add() canonicalizes keys but the vegeta API is used to test systems that require case-sensitive headers.
		header[key] = append(header[key], val)
	}

	if c.body != "" && c.bodyFile != "" {
		return nil, fmt.Errorf("only one of --body and --body-file can be specified")
	}

	body := []byte(c.body)
	if c.bodyFile != "" {
		if b, err := ioutil.ReadFile(c.bodyFile); err != nil {
			return nil, fmt.Errorf("unable to open %q: %w", c.bodyFile, err)
		} else {
			body = b
		}
	}

	return &attacker.Options{
		Rate:     c.rate,
		Duration: c.duration,
		Timeout:  c.timeout,
		Method:   c.method,
		Body:     body,
		Header:   header,
	}, nil
}

func validateMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace:
		return true
	}
	return false
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
