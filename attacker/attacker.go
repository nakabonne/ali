package attacker

import (
	"context"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

const (
	DefaultRate        = 50
	DefaultDuration    = 10 * time.Second
	DefaultTimeout     = 30 * time.Second
	DefaultMethod      = http.MethodGet
	DefaultWorkers     = 10
	DefaultMaxWorkers  = math.MaxUint64
	DefaultMaxBody     = int64(-1)
	DefaultConnections = 10000
)

var DefaultLocalAddr = net.IPAddr{IP: net.IPv4zero}

type Attacker interface {
	Attack(vegeta.Targeter, vegeta.Pacer, time.Duration, string) <-chan *vegeta.Result
	Stop()
}

// Options provides optional settings to attack.
type Options struct {
	Rate        int
	Duration    time.Duration
	Timeout     time.Duration
	Method      string
	Body        []byte
	MaxBody     int64
	Header      http.Header
	Workers     uint64
	MaxWorkers  uint64
	KeepAlive   bool
	Connections int
	HTTP2       bool
	LocalAddr   net.IPAddr
	Buckets     []time.Duration
	Resolvers   []string
	Variables   [][]string
	Attacker    Attacker
}

// Result contains the results of a single HTTP request.
type Result struct {
	Latency time.Duration

	P50 time.Duration
	P90 time.Duration
	P95 time.Duration
	P99 time.Duration
}

// Attack keeps the request running for the specified period of time.
// Results are sent to the given channel as soon as they arrive.
// When the attack is over, it gives back final statistics.
func Attack(ctx context.Context, target string, resCh chan<- *Result, metricsCh chan *Metrics, opts Options) {
	if target == "" {
		return
	}
	if opts.Method == "" {
		opts.Method = DefaultMethod
	}
	if opts.Workers == 0 {
		opts.Workers = DefaultWorkers
	}
	if opts.MaxWorkers == 0 {
		opts.MaxWorkers = DefaultMaxWorkers
	}
	if opts.MaxBody == 0 {
		opts.MaxBody = DefaultMaxBody
	}
	if opts.Connections == 0 {
		opts.Connections = DefaultConnections
	}
	if opts.LocalAddr.IP == nil {
		opts.LocalAddr = DefaultLocalAddr
	}
	if len(opts.Resolvers) > 0 {
		net.DefaultResolver = NewResolver(opts.Resolvers)
	}
	if opts.Attacker == nil {
		opts.Attacker = vegeta.NewAttacker(
			vegeta.Timeout(opts.Timeout),
			vegeta.Workers(opts.Workers),
			vegeta.MaxWorkers(opts.MaxWorkers),
			vegeta.MaxBody(opts.MaxBody),
			vegeta.Connections(opts.Connections),
			vegeta.KeepAlive(opts.KeepAlive),
			vegeta.HTTP2(opts.HTTP2),
			vegeta.LocalAddr(opts.LocalAddr),
		)
	}

	var targeter vegeta.Targeter
	rate := vegeta.Rate{Freq: opts.Rate, Per: time.Second}
	if opts.Variables == nil {
		targeter = vegeta.NewStaticTargeter(vegeta.Target{
			Method: opts.Method,
			URL:    target,
			Body:   opts.Body,
			Header: opts.Header,
		})
	} else {
		tmpURLfile, err := ioutil.TempFile("", "ali-")
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(tmpURLfile.Name())
		buildTmpURLFile(tmpURLfile, target, int(opts.Duration/time.Second)*opts.Rate, opts.Variables)
		targeter = vegeta.NewHTTPTargeter(tmpURLfile, opts.Body, opts.Header)
	}

	metrics := &vegeta.Metrics{}
	if len(opts.Buckets) > 0 {
		metrics.Histogram = &vegeta.Histogram{Buckets: opts.Buckets}
	}

	for res := range opts.Attacker.Attack(targeter, rate, opts.Duration, "main") {
		select {
		case <-ctx.Done():
			opts.Attacker.Stop()
			return
		default:
			metrics.Add(res)
			m := newMetrics(metrics)
			resCh <- &Result{
				Latency: res.Latency,
				P50:     m.Latencies.P50,
				P90:     m.Latencies.P90,
				P95:     m.Latencies.P95,
				P99:     m.Latencies.P99,
			}
			metricsCh <- m
		}
	}
	metrics.Close()
	metricsCh <- newMetrics(metrics)
}

func buildTmpURLFile(tmpURLfile *os.File, target string, reqCount int, variables [][]string) {
	pattern := regexp.MustCompile(`\$\{\d{1,4}\}`)
	for z := 0; z < reqCount; z++ {
		tmpUrl := target
		for _, submatches := range pattern.FindAllSubmatchIndex([]byte(target), -1) {
			i, _ := strconv.Atoi(target[submatches[0]+2 : submatches[1]-1])
			if len(variables) >= i {
				rndValue := variables[i-1][rand.Intn(len(variables[i-1]))]
				tmpUrl = strings.ReplaceAll(tmpUrl, target[submatches[0]:submatches[1]], rndValue)
			}
		}
		tmpURLfile.WriteString(tmpUrl + "\n")
	}
	tmpURLfile.Seek(0, 0)
}
