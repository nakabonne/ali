package gui

import (
	"context"
	"math"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/textinput"
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
	var inputs []float64
	for i := 0; i < 100; i++ {
		v := math.Pow(math.Sin(float64(i)), 63) * math.Sin(float64(i)+1.5) * 8
		inputs = append(inputs, v)
	}

	lc, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
	)
	if err != nil {
		return nil, err
	}
	step := 0
	go periodic(ctx, redrawInterval/3, func() error {
		step = (step + 1) % len(inputs)
		return lc.Series("heartbeat", rotateFloats(inputs, step),
			linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(87))),
			linechart.SeriesXLabels(map[int]string{
				0: "zero",
			}),
		)
	})
	return lc, nil
}

// periodic executes the provided closure periodically every interval.
// Exits when the context expires.
func periodic(ctx context.Context, interval time.Duration, fn func() error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := fn(); err != nil {
				panic(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// rotateFloats returns a new slice with inputs rotated by step.
// I.e. for a step of one:
//   inputs[0] -> inputs[len(inputs)-1]
//   inputs[1] -> inputs[0]
// And so on.
func rotateFloats(inputs []float64, step int) []float64 {
	return append(inputs[step:], inputs[:step]...)
}
