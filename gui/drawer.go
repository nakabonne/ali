package gui

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
	"go.uber.org/atomic"

	"github.com/nakabonne/ali/attacker"
	"github.com/nakabonne/ali/storage"
)

// drawer periodically queries data points from the storage and passes them to the termdash API.
type drawer struct {
	// specify the data points range to show on the UI
	queryRange     time.Duration
	redrawInterval time.Duration
	widgets        *widgets
	gridOpts       *gridOpts

	metricsCh chan *attacker.Metrics

	// aims to avoid to perform multiple `appendChartValues`.
	chartDrawing *atomic.Bool

	mu      sync.RWMutex
	metrics *attacker.Metrics
	storage storage.Reader
}

// redrawCharts sets the values held by itself as chart values, at the specified interval as redrawInterval.
func (d *drawer) redrawCharts(ctx context.Context) {
	ticker := time.NewTicker(d.redrawInterval)
	defer ticker.Stop()

	d.chartDrawing.Store(true)
L:
	for {
		select {
		case <-ctx.Done():
			break L
		case <-ticker.C:
			end := time.Now()
			start := end.Add(-d.queryRange)

			latencies, err := d.storage.Select(storage.LatencyMetricName, start, end)
			if err != nil {
				log.Printf("failed to select latency data points: %v\n", err)
			}
			d.widgets.latencyChart.Series("latency", latencies,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(87))),
				linechart.SeriesXLabels(map[int]string{
					0: "req",
				}),
			)

			p50, err := d.storage.Select(storage.P50MetricName, start, end)
			if err != nil {
				log.Printf("failed to select p50 data points: %v\n", err)
			}
			d.widgets.percentilesChart.Series("p50", p50,
				linechart.SeriesCellOpts(d.widgets.p50Legend.cellOpts...),
			)

			p90, err := d.storage.Select(storage.P90MetricName, start, end)
			if err != nil {
				log.Printf("failed to select p90 data points: %v\n", err)
			}
			d.widgets.percentilesChart.Series("p90", p90,
				linechart.SeriesCellOpts(d.widgets.p90Legend.cellOpts...),
			)

			p95, err := d.storage.Select(storage.P95MetricName, start, end)
			if err != nil {
				log.Printf("failed to select p95 data points: %v\n", err)
			}
			d.widgets.percentilesChart.Series("p95", p95,
				linechart.SeriesCellOpts(d.widgets.p95Legend.cellOpts...),
			)

			p99, err := d.storage.Select(storage.P99MetricName, start, end)
			if err != nil {
				log.Printf("failed to select p99 data points: %v\n", err)
			}
			d.widgets.percentilesChart.Series("p99", p99,
				linechart.SeriesCellOpts(d.widgets.p99Legend.cellOpts...),
			)
		}
	}
	d.chartDrawing.Store(false)
}

func (d *drawer) redrawGauge(ctx context.Context, duration time.Duration) {
	ticker := time.NewTicker(d.redrawInterval)
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
	ticker := time.NewTicker(d.redrawInterval)
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
