package gui

import (
	"context"
	"time"

	"github.com/mum4k/termdash/widgets/text"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/linechart"

	"github.com/nakabonne/ali/attacker"
)

type drawer struct {
	widgets  *widgets
	chartsCh chan *attacker.Result
	reportCh chan string
}

// TODO: In the future, multiple charts including bytes-in/out etc will be re-drawn.
func (d *drawer) redrawChart(ctx context.Context, maxSize int) {
	values := make([]float64, 0, maxSize)
	for {
		select {
		case <-ctx.Done():
			return
		case res := <-d.chartsCh:
			values = append(values, float64(res.Latency/time.Millisecond))
			d.widgets.latencyChart.Series("latency", values,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(87))),
				linechart.SeriesXLabels(map[int]string{
					0: "req",
				}),
			)
		}
	}
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
