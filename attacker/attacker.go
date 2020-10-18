package attacker

import (
	"context"
	"math"
	"net"
	"net/http"
	"sync"
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

	Attacker Attacker
}

// Result contains the results of a single HTTP request.
type Result struct {
	Latency time.Duration

	P50 time.Duration
	P90 time.Duration
	P95 time.Duration
	P99 time.Duration

	// Indicates if the last result in the entire attack.
	End bool
}

// Attack keeps the request running for the specified period of time.
// Results are sent to the given channel as soon as they arrive.
// When the attack is over, it gives back final statistics.
func Attack(ctx context.Context, target string, resCh chan *Result, metricsCh chan *Metrics, opts Options) {
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

	rate := vegeta.Rate{Freq: opts.Rate, Per: time.Second}
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: opts.Method,
		URL:    target,
		Body:   opts.Body,
		Header: opts.Header,
	})

	metrics := &vegeta.Metrics{}

	child, cancelChild := context.WithCancel(ctx)
	defer cancelChild()

	// used to protect metrics
	mu := &sync.Mutex{}
	go sendMetrics(child, metrics, metricsCh, mu)

	for res := range opts.Attacker.Attack(targeter, rate, opts.Duration, "main") {
		select {
		case <-ctx.Done():
			opts.Attacker.Stop()
			return
		default:
			mu.Lock()
			metrics.Add(res)
			p50 := metrics.Latencies.Quantile(0.50)
			p90 := metrics.Latencies.Quantile(0.90)
			p95 := metrics.Latencies.Quantile(0.95)
			p99 := metrics.Latencies.Quantile(0.99)
			mu.Unlock()
			resCh <- &Result{
				Latency: res.Latency,
				P50:     p50,
				P90:     p90,
				P95:     p95,
				P99:     p99,
			}
		}
	}
	metrics.Close()
	metricsCh <- newMetrics(metrics, mu)
}

func sendMetrics(ctx context.Context, metrics *vegeta.Metrics, ch chan<- *Metrics, mu *sync.Mutex) {
	// TODO: Make the interval changeable.
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ch <- newMetrics(metrics, mu)
		}
	}
}
