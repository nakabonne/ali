package gui

import (
	"context"
	"time"

	"github.com/k0kubun/pp"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/textinput"

	"github.com/nakabonne/ali/attacker"
)

// redrawInterval is how often termdash redraws the screen.
const (
	redrawInterval = 250 * time.Millisecond
)

type widgets struct {
	URLInput     *textinput.TextInput
	attackButton *button.Button
	plotChart    *linechart.LineChart
}

func newWidgets(ctx context.Context, c *container.Container) (*widgets, error) {
	l, err := newLineChart(ctx)
	if err != nil {
		return nil, err
	}

	return &widgets{
		URLInput:     nil,
		attackButton: nil,
		plotChart:    l,
	}, nil
}

// newLineChart returns a line plotChart that displays a heartbeat-like progression.
func newLineChart(ctx context.Context) (*linechart.LineChart, error) {
	lc, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
	)
	if err != nil {
		return nil, err
	}

	resultCh := make(chan *attacker.Result)
	go func() {
		// TODO: Enalble to poplulate from input
		metrics := attacker.Attack(ctx, "http://34.84.111.163:9898", resultCh, attacker.Options{Rate: 50, Duration: 10 * time.Second})
		pp.Println(metrics)
	}()
	go redrawChart(ctx, lc, resultCh)

	return lc, nil
}

func redrawChart(ctx context.Context, lineChart *linechart.LineChart, resultCh chan *attacker.Result) {
	values := []float64{}
	for {
		select {
		case <-ctx.Done():
			return
		case res := <-resultCh:
			pp.Println("latency", res.Latency/time.Millisecond)
			values = append(values, float64(res.Latency/time.Millisecond))
			lineChart.Series("plot", values,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(87))),
				linechart.SeriesXLabels(map[int]string{
					0: "num",
				}),
			)
		}
	}
}
