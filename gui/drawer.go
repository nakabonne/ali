package gui

import (
	"context"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"

	"github.com/nakabonne/ali/attacker"
)

type drawer struct {
	widgets  *widgets
	chartsCh chan *attacker.Result
	reportCh chan string

	// Indicates if charts are drawing.
	chartDrawing bool
}

// redrawChart appends entities as soon as a result arrives.
// Given maxSize, then it can be pre-allocated.
// TODO: In the future, multiple charts including bytes-in/out etc will be re-drawn.
func (d *drawer) redrawChart(ctx context.Context, maxSize int) {
	values := make([]float64, 0, maxSize)
	d.chartDrawing = true
L:
	for {
		select {
		case <-ctx.Done():
			break L
		case res := <-d.chartsCh:
			if res.End {
				break L
			}
			values = append(values, float64(res.Latency/time.Millisecond))
			d.widgets.latencyChart.Series("latency", values,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(87))),
				linechart.SeriesXLabels(map[int]string{
					0: "req",
				}),
			)
		}
	}
	d.chartDrawing = false
}

func (d *drawer) redrawReport(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case report := <-d.reportCh:
			d.widgets.reportText.Write(report, text.WriteReplace())
		}
	}
}
