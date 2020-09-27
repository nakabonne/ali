package attacker

import (
	"context"
	"math"
	"net/http"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

const (
	DefaultRate       = 50
	DefaultDuration   = 10 * time.Second
	DefaultTimeout    = 30 * time.Second
	DefaultMethod     = http.MethodGet
	DefaultWorkers    = 10
	DefaultMaxWorkers = math.MaxUint64
)

type Attacker interface {
	Attack(vegeta.Targeter, vegeta.Pacer, time.Duration, string) <-chan *vegeta.Result
	Stop()
}

// Options provides optional settings to attack.
type Options struct {
	Rate       int
	Duration   time.Duration
	Timeout    time.Duration
	Method     string
	Body       []byte
	Header     http.Header
	Workers    uint64
	MaxWorkers uint64

	Attacker Attacker
}

// Result contains the results of a single HTTP request.
type Result struct {
	Latency time.Duration
	// Indicates if the last result in the entire attack.
	End bool
}

// Attack keeps the request running for the specified period of time.
// Results are sent to the given channel as soon as they arrive.
// When the attack is over, it gives back final statistics.
func Attack(ctx context.Context, target string, resCh chan *Result, opts Options) *Metrics {
	if target == "" {
		return nil
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
	if opts.Attacker == nil {
		opts.Attacker = vegeta.NewAttacker(
			vegeta.Timeout(opts.Timeout),
			vegeta.Workers(opts.Workers),
			vegeta.MaxWorkers(opts.MaxWorkers),
		)
	}

	rate := vegeta.Rate{Freq: opts.Rate, Per: time.Second}
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: opts.Method,
		URL:    target,
		Body:   opts.Body,
		Header: opts.Header,
	})

	var metrics vegeta.Metrics

	for res := range opts.Attacker.Attack(targeter, rate, opts.Duration, "main") {
		select {
		case <-ctx.Done():
			opts.Attacker.Stop()
		default:
			resCh <- &Result{Latency: res.Latency}
			metrics.Add(res)
		}
	}
	metrics.Close()

	return newMetrics(&metrics)
}
