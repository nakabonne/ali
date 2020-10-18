package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
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

	// Automatically populated by goreleaser during build
	version = "unversioned"
	commit  = "?"
	date    = "?"
)

type cli struct {
	rate         int
	duration     time.Duration
	timeout      time.Duration
	method       string
	headers      []string
	body         string
	bodyFile     string
	maxBody      int64
	workers      uint64
	maxWorkers   uint64
	connections  int
	noHTTP2      bool
	localAddress string
	noKeepAlive  bool
	buckets      string

	debug   bool
	version bool
	stdout  io.Writer
	stderr  io.Writer
}

func main() {
	c, err := parseFlags(os.Stdout, os.Stderr)
	if err != nil {
		os.Exit(0)
	}
	os.Exit(c.run(flagSet.Args()))
}

func parseFlags(stdout, stderr io.Writer) (*cli, error) {
	c := &cli{
		stdout: stdout,
		stderr: stderr,
	}
	flagSet.IntVarP(&c.rate, "rate", "r", attacker.DefaultRate, "The request rate per second to issue against the targets. Give 0 then it will send requests as fast as possible.")
	flagSet.DurationVarP(&c.duration, "duration", "d", attacker.DefaultDuration, "The amount of time to issue requests to the targets. Give 0s for an infinite attack.")
	flagSet.DurationVarP(&c.timeout, "timeout", "t", attacker.DefaultTimeout, "The timeout for each request. 0s means to disable timeouts.")
	flagSet.StringVarP(&c.method, "method", "m", attacker.DefaultMethod, "An HTTP request method for each request.")
	flagSet.StringSliceVarP(&c.headers, "header", "H", []string{}, "A request header to be sent. Can be used multiple times to send multiple headers.")
	flagSet.StringVarP(&c.body, "body", "b", "", "A request body to be sent.")
	flagSet.StringVarP(&c.bodyFile, "body-file", "B", "", "The path to file whose content will be set as the http request body.")
	flagSet.Int64VarP(&c.maxBody, "max-body", "M", attacker.DefaultMaxBody, "Max bytes to capture from response bodies. Give -1 for no limit.")
	flagSet.BoolVarP(&c.version, "version", "v", false, "Print the current version.")
	flagSet.BoolVar(&c.debug, "debug", false, "Run in debug mode.")
	flagSet.BoolVarP(&c.noKeepAlive, "no-keepalive", "K", false, "Don't use HTTP persistent connection.")
	flagSet.Uint64VarP(&c.workers, "workers", "w", attacker.DefaultWorkers, "Amount of initial workers to spawn.")
	flagSet.Uint64VarP(&c.maxWorkers, "max-workers", "W", attacker.DefaultMaxWorkers, "Amount of maximum workers to spawn.")
	flagSet.IntVarP(&c.connections, "connections", "c", attacker.DefaultConnections, "Amount of maximum open idle connections per target host")
	flagSet.BoolVar(&c.noHTTP2, "no-http2", false, "Don't issue HTTP/2 requests to servers which support it.")
	flagSet.StringVar(&c.localAddress, "local-addr", "0.0.0.0", "Local IP address.")
	// TODO: Re-enable when making it capable of drawing histogram bar chart.
	//flagSet.StringVar(&c.buckets, "buckets", "", "Histogram buckets; comma-separated list.")
	flagSet.Usage = c.usage
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(c.stderr, err)
		}
		return nil, err
	}
	return c, nil
}

func (c *cli) run(args []string) int {
	if c.version {
		fmt.Fprintf(c.stderr, "version=%s, commit=%s, buildDate=%s, os=%s, arch=%s\n", version, commit, date, runtime.GOOS, runtime.GOARCH)
		return 0
	}
	if len(args) == 0 {
		fmt.Fprintln(c.stderr, "no target given")
		c.usage()
		return 1
	}
	target := args[0]
	if _, err := url.ParseRequestURI(target); err != nil {
		fmt.Fprintf(c.stderr, "bad target URL: %v\n", err)
		c.usage()
		return 1
	}
	opts, err := c.makeOptions()
	if err != nil {
		fmt.Fprintln(c.stderr, err.Error())
		c.usage()
		return 1
	}
	setDebug(nil, c.debug)
	if err := gui.Run(target, opts); err != nil {
		fmt.Fprintf(c.stderr, "failed to start application: %s\n", err.Error())
		c.usage()
		return 1
	}

	return 0
}

func (c *cli) usage() {
	format := `Usage:
  ali [flags] <target URL>

Flags:
%s
Examples:
  ali --duration=10m --rate=100 http://host.xz

Author:
  Ryo Nakao <ryo@nakao.dev>
`
	fmt.Fprintf(c.stderr, format, flagSet.FlagUsages())
}

// makeOptions gives back an options for attacker, with the CLI input.
func (c *cli) makeOptions() (*attacker.Options, error) {
	if !validateMethod(c.method) {
		return nil, fmt.Errorf("given method %q isn't an HTTP request method", c.method)
	}
	if c.duration <= 0 {
		return nil, fmt.Errorf("duration must be greater than 0")
	}

	header := make(http.Header)
	for _, hdr := range c.headers {
		parts := strings.SplitN(hdr, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("given header %q has a wrong format", hdr)
		}
		key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if key == "" || val == "" {
			return nil, fmt.Errorf("given header %q has a wrong format", hdr)
		}
		// NOTE: Add key/value directly to the http.Header (map[string][]string).
		// http.Header.Add() canonicalizes keys but the vegeta API is used to test systems that require case-sensitive headers.
		header[key] = append(header[key], val)
	}

	if c.body != "" && c.bodyFile != "" {
		return nil, fmt.Errorf(`only one of "--body" and "--body-file" can be specified`)
	}

	body := []byte(c.body)
	if c.bodyFile != "" {
		b, err := ioutil.ReadFile(c.bodyFile)
		if err != nil {
			return nil, fmt.Errorf("unable to open %q: %w", c.bodyFile, err)
		}
		body = b
	}

	localAddr := net.IPAddr{IP: net.ParseIP(c.localAddress)}

	parsedBuckets, err := parseBucketOptions(c.buckets)

	if err != nil {
		return nil, fmt.Errorf("wrong buckets format %w", err)
	}

	return &attacker.Options{
		Rate:        c.rate,
		Duration:    c.duration,
		Timeout:     c.timeout,
		Method:      c.method,
		Body:        body,
		MaxBody:     c.maxBody,
		Header:      header,
		KeepAlive:   !c.noKeepAlive,
		Workers:     c.workers,
		MaxWorkers:  c.maxWorkers,
		Connections: c.connections,
		HTTP2:       !c.noHTTP2,
		LocalAddr:   localAddr,
		Buckets:     parsedBuckets,
	}, nil
}

func validateMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace:
		return true
	}
	return false
}

func parseBucketOptions(rawBuckets string) ([]time.Duration, error) {
	if rawBuckets == "" {
		return []time.Duration{}, nil
	}

	stringBuckets := strings.Split(rawBuckets, ",")
	result := make([]time.Duration, len(stringBuckets))

	for _, bucket := range stringBuckets {
		trimmedBucket := strings.TrimSpace(bucket)
		d, err := time.ParseDuration(trimmedBucket)
		if err != nil {
			return nil, err
		}
		result = append(result, d)
	}

	return result, nil
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
