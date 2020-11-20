package gui

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
	"go.uber.org/atomic"

	"github.com/nakabonne/ali/attacker"
)

// drawer buffers the result values because calling the termdash API
// whenever itreceives a result would create a bottleneck.
// To draw them, it periodically passes those values to the termdash API.
type drawer struct {
	widgets  *widgets
	gridOpts *gridOpts

	chartCh   chan *attacker.Result
	metricsCh chan *attacker.Metrics
	doneCh    chan struct{}

	// aims to avoid to perform multiple `appendChartValues`.
	chartDrawing *atomic.Bool

	mu          sync.RWMutex
	chartValues values
	metrics     *attacker.Metrics
}

type values struct {
	latencies []float64
	p50       []float64
	p90       []float64
	p95       []float64
	p99       []float64
}

// appendChartValues appends entities as soon as a result arrives.
// Given maxSize, then it can be pre-allocated.
func (d *drawer) appendChartValues(ctx context.Context, rate int, duration time.Duration) {
	// TODO: Change how to stop `redrawGauge`.
	// We currently use this way to ensure to stop `redrawGauge` after the increase process is complete.
	// But, it's preferable to stop goroutine where it's generated.
	maxSize := rate * int(duration/time.Second)
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	go d.redrawGauge(child, duration)

	d.chartValues.latencies = make([]float64, 0, maxSize)
	d.chartValues.p50 = make([]float64, 0, maxSize)
	d.chartValues.p90 = make([]float64, 0, maxSize)
	d.chartValues.p95 = make([]float64, 0, maxSize)
	d.chartValues.p99 = make([]float64, 0, maxSize)

	appendValue := func(to []float64, val time.Duration) []float64 {
		return append(to, float64(val)/float64(time.Millisecond))
	}

L:
	for {
		select {
		case <-ctx.Done():
			break L
		case <-d.doneCh:
			break L
		case res := <-d.chartCh:
			if res == nil {
				continue
			}

			d.mu.Lock()
			d.chartValues.latencies = appendValue(d.chartValues.latencies, res.Latency)
			d.chartValues.p50 = appendValue(d.chartValues.p50, res.P50)
			d.chartValues.p90 = appendValue(d.chartValues.p90, res.P90)
			d.chartValues.p95 = appendValue(d.chartValues.p95, res.P95)
			d.chartValues.p99 = appendValue(d.chartValues.p99, res.P99)
			d.mu.Unlock()
		}
	}
}

// redrawCharts sets the values held by itself as chart values, at the specified interval as redrawInterval.
func (d *drawer) redrawCharts(ctx context.Context) {
	ticker := time.NewTicker(redrawInterval)
	defer ticker.Stop()

	d.chartDrawing.Store(true)
L:
	for {
		select {
		case <-ctx.Done():
			break L
		case <-d.doneCh:
			break L
		case <-ticker.C:
			d.widgets.latencyChart.Series("latency", d.chartValues.latencies,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(87))),
				linechart.SeriesXLabels(map[int]string{
					0: "req",
				}),
			)
			d.widgets.percentilesChart.Series("p50", d.chartValues.p50,
				linechart.SeriesCellOpts(d.widgets.p50Legend.cellOpts...),
			)
			d.widgets.percentilesChart.Series("p90", d.chartValues.p90,
				linechart.SeriesCellOpts(d.widgets.p90Legend.cellOpts...),
			)
			d.widgets.percentilesChart.Series("p95", d.chartValues.p95,
				linechart.SeriesCellOpts(d.widgets.p95Legend.cellOpts...),
			)
			d.widgets.percentilesChart.Series("p99", d.chartValues.p99,
				linechart.SeriesCellOpts(d.widgets.p99Legend.cellOpts...),
			)
		}
	}
	d.chartDrawing.Store(false)
}

func (d *drawer) redrawGauge(ctx context.Context, duration time.Duration) {
	ticker := time.NewTicker(redrawInterval)
	defer ticker.Stop()

	totalTime := float64(duration)

	d.widgets.progressGauge.Percent(0)
	for start := time.Now(); ; {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			passed := float64(time.Since(start))
			percent := int(passed / totalTime * 100)
			// as time.Duration is the unit of nanoseconds
			// small duration can exceed 100 on slow machines
			if percent > 100 {
				continue
			}
			d.widgets.progressGauge.Percent(percent)
		}
	}
}

const (
	latenciesTextFormat = `Total: %v
Mean: %v
P50: %v
P90: %v
P95: %v
P99: %v
Max: %v
Min: %v`

	bytesTextFormat = `In:
  Total: %v
  Mean: %v
Out:
  Total: %v
  Mean: %v`

	othersTextFormat = `Duration: %v
Wait: %v
Requests: %d
Rate: %f
Throughput: %f
Success: %f
Earliest: %v
Latest: %v
End: %v`
)

// redrawMetrics writes the metrics held by itself into the widgets, at the specified interval as redrawInterval.
func (d *drawer) redrawMetrics(ctx context.Context) {
	ticker := time.NewTicker(redrawInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.mu.RLock()
			m := *d.metrics
			d.mu.RUnlock()

			d.widgets.latenciesText.Write(
				fmt.Sprintf(latenciesTextFormat,
					m.Latencies.Total,
					m.Latencies.Mean,
					m.Latencies.P50,
					m.Latencies.P90,
					m.Latencies.P95,
					m.Latencies.P99,
					m.Latencies.Max,
					m.Latencies.Min,
				), text.WriteReplace())

			d.widgets.bytesText.Write(
				fmt.Sprintf(bytesTextFormat,
					m.BytesIn.Total,
					m.BytesIn.Mean,
					m.BytesOut.Total,
					m.BytesOut.Mean,
				), text.WriteReplace())

			d.widgets.othersText.Write(fmt.Sprintf(othersTextFormat,
				m.Duration,
				m.Wait,
				m.Requests,
				m.Rate,
				m.Throughput,
				m.Success,
				m.Earliest.Format(time.RFC3339),
				m.Latest.Format(time.RFC3339),
				m.End.Format(time.RFC3339),
			), text.WriteReplace())

			// To guarantee that status codes are in order
			// taking the slice of keys and sorting them.
			codesText := ""
			var keys []string
			for k := range m.StatusCodes {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				codesText += fmt.Sprintf(`%q: %d
`, k, m.StatusCodes[k])
			}
			d.widgets.statusCodesText.Write(codesText, text.WriteReplace())

			errorsText := ""
			for _, e := range m.Errors {
				errorsText += fmt.Sprintf(`- %s
`, e)
			}
			d.widgets.errorsText.Write(errorsText, text.WriteReplace())
		}
	}
}

func (d *drawer) updateMetrics(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case metrics := <-d.metricsCh:
			if metrics == nil {
				continue
			}
			d.mu.Lock()
			d.metrics = metrics
			d.mu.Unlock()
		}
	}
}
