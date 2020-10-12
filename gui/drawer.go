package gui

import (
	"context"
	"fmt"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"

	"github.com/nakabonne/ali/attacker"
)

type drawer struct {
	widgets   *widgets
	gridOpts  *gridOpts
	chartCh   chan *attacker.Result
	gaugeCh   chan bool
	metricsCh chan *attacker.Metrics

	// aims to avoid to perform multiple `redrawChart`.
	chartDrawing bool
}

// redrawChart appends entities as soon as a result arrives.
// Given maxSize, then it can be pre-allocated.
// TODO: In the future, multiple charts including bytes-in/out etc will be re-drawn.
func (d *drawer) redrawChart(ctx context.Context, maxSize int) {
	values := make([]float64, 0, maxSize)

	valuesP50 := make([]float64, 0, maxSize)
	valuesP90 := make([]float64, 0, maxSize)
	valuesP95 := make([]float64, 0, maxSize)
	valuesP99 := make([]float64, 0, maxSize)

	appendValue := func(to []float64, val time.Duration) []float64 {
		return append(to, float64(val)/float64(time.Millisecond))
	}

	d.chartDrawing = true
L:
	for {
		select {
		case <-ctx.Done():
			break L
		case res := <-d.chartCh:
			if res == nil {
				continue
			}
			if res.End {
				d.gaugeCh <- true
				break L
			}
			d.gaugeCh <- false

			values = appendValue(values, res.Latency)
			d.widgets.latencyChart.Series("latency", values,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(87))),
				linechart.SeriesXLabels(map[int]string{
					0: "req",
				}),
			)

			valuesP50 = appendValue(valuesP50, res.P50)
			d.widgets.percentilesChart.Series("p50", valuesP50,
				linechart.SeriesCellOpts(d.widgets.p50Legend.cellOpts...),
			)

			valuesP90 = appendValue(valuesP90, res.P90)
			d.widgets.percentilesChart.Series("p90", valuesP90,
				linechart.SeriesCellOpts(d.widgets.p90Legend.cellOpts...),
			)

			valuesP95 = appendValue(valuesP95, res.P95)
			d.widgets.percentilesChart.Series("p95", valuesP95,
				linechart.SeriesCellOpts(d.widgets.p95Legend.cellOpts...),
			)

			valuesP99 = appendValue(valuesP99, res.P99)
			d.widgets.percentilesChart.Series("p99", valuesP99,
				linechart.SeriesCellOpts(d.widgets.p99Legend.cellOpts...),
			)
		}
	}
	d.chartDrawing = false
}

func (d *drawer) redrawGauge(ctx context.Context, maxSize int) {
	var count float64
	size := float64(maxSize)
	d.widgets.progressGauge.Percent(0)
	for {
		select {
		case <-ctx.Done():
			return
		case end := <-d.gaugeCh:
			if end {
				return
			}
			count++
			d.widgets.progressGauge.Percent(int(count / size * 100))
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

func (d *drawer) redrawMetrics(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case metrics := <-d.metricsCh:
			if metrics == nil {
				continue
			}
			d.widgets.latenciesText.Write(
				fmt.Sprintf(latenciesTextFormat,
					metrics.Latencies.Total,
					metrics.Latencies.Mean,
					metrics.Latencies.P50,
					metrics.Latencies.P90,
					metrics.Latencies.P95,
					metrics.Latencies.P99,
					metrics.Latencies.Max,
					metrics.Latencies.Min,
				), text.WriteReplace())

			d.widgets.bytesText.Write(
				fmt.Sprintf(bytesTextFormat,
					metrics.BytesIn.Total,
					metrics.BytesIn.Mean,
					metrics.BytesOut.Total,
					metrics.BytesOut.Mean,
				), text.WriteReplace())

			d.widgets.othersText.Write(fmt.Sprintf(othersTextFormat,
				metrics.Duration,
				metrics.Wait,
				metrics.Requests,
				metrics.Rate,
				metrics.Throughput,
				metrics.Success,
				metrics.Earliest.Format(time.RFC3339),
				metrics.Latest.Format(time.RFC3339),
				metrics.End.Format(time.RFC3339),
			), text.WriteReplace())

			codesText := ""
			for code, n := range metrics.StatusCodes {
				codesText += fmt.Sprintf(`%q: %d
`, code, n)
			}
			d.widgets.statusCodesText.Write(codesText, text.WriteReplace())

			errorsText := ""
			for _, e := range metrics.Errors {
				errorsText += fmt.Sprintf(`- %s
`, e)
			}
			d.widgets.errorsText.Write(errorsText, text.WriteReplace())
		}
	}
}
