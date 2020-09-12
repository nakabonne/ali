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

type Attacker interface {
	Attack(vegeta.Targeter, vegeta.Pacer, time.Duration, string) <-chan *vegeta.Result
	Stop()
}

// Options provides optional settings to attack.
type Options struct {
	Rate     int
	Duration time.Duration
	Method   string
	Body     []byte
	Header   http.Header
	Attacker Attacker
}

// Result contains the results of a single Target hit.
type Result struct {
	Latency time.Duration
}

// Attack keeps the request running for the specified period of time.
// Results are sent to the given channel as soon as they arrive.
// When the attack is over, it gives back final statistics.
func Attack(ctx context.Context, target string, resCh chan *Result, opts Options) *Metrics {
	if target == "" {
		return nil
	}
	if opts.Rate == 0 {
		opts.Rate = defaultRate
	}
	if opts.Duration == 0 {
		opts.Duration = defaultDuration
	}
	if opts.Method == "" {
		opts.Method = defaultMethod
	}
	if opts.Attacker == nil {
		opts.Attacker = vegeta.NewAttacker()
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
