package attacker

import (
	"context"
	"net/http"
	"time"

	"github.com/k0kubun/pp"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

const (
	defaultRate     = 50
	defaultDuration = 30 * time.Second
	defaultMethod   = http.MethodGet
)

// Options provides optional settings to attack.
type Options struct {
	Rate     int
	Duration time.Duration
	Method   string
	Body     []byte
	Header   http.Header
	// TODO: Make attacker pluggable to make it easy unit tests.
	//Attacker func(vegeta.Targeter, vegeta.Pacer, time.Duration, string) <-chan *vegeta.Result
}

// Result contains the results of a single Target hit.
type Result struct {
	Latency time.Duration
}

// Attack keeps the request running for the specified period of time.
// Results are sent to the given channel as soon as they arrive.
// When the attack is over, it gives back final statistics.
func Attack(ctx context.Context, target string, resCh chan *Result, opts Options) *Metrics {
	pp.Println("start attacking")
	if opts.Rate == 0 {
		opts.Rate = defaultRate
	}
	if opts.Duration == 0 {
		opts.Duration = defaultDuration
	}
	if opts.Method == "" {
		opts.Method = defaultMethod
	}

	rate := vegeta.Rate{Freq: opts.Rate, Per: time.Second}
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: opts.Method,
		URL:    target,
		Body:   opts.Body,
		Header: opts.Header,
	})
	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics

	for res := range attacker.Attack(targeter, rate, opts.Duration, "main") {
		// TODO: Stop if ctx is expired.
		/*if <-ctx.Done() {
			return newMetrics(&metrics)
		}*/
		resCh <- &Result{Latency: res.Latency}
		metrics.Add(res)
	}
	metrics.Close()
	pp.Println("vegeta metrics", metrics)

	return newMetrics(&metrics)
}

type (
	// Metrics wraps vegeta.Metrics to avoid dependency on it.
	Metrics struct {
		Latencies LatencyMetrics
	}
	LatencyMetrics struct {
		Total time.Duration
		Mean  time.Duration
		P50   time.Duration
		P90   time.Duration
		P95   time.Duration
		P99   time.Duration
		Max   time.Duration
		Min   time.Duration
	}
)

func newMetrics(m *vegeta.Metrics) *Metrics {
	return &Metrics{
		Latencies: LatencyMetrics{
			Total: m.Latencies.Total,
			Mean:  m.Latencies.Mean,
			P50:   m.Latencies.P50,
			P90:   m.Latencies.P90,
			P95:   m.Latencies.P95,
			P99:   m.Latencies.P99,
			Max:   m.Latencies.Max,
			Min:   m.Latencies.Min,
		},
	}
}
